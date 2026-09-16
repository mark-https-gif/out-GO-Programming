//go:build windows

package stdlib

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/out-lang/out/internal/object"
)

// Win32 constants used for raw device access.
const (
	genericRead         = 0x80000000
	genericWrite        = 0x40000000
	fileFlagWriteThrough = 0x80000000 // overrides buffering, writes straight to disk
	fileFlagNoBuffering  = 0x20000000
	fileFlagSequentialScan = 0x08000000
	openExisting        = 3
	fileShareRead       = 1
	fileShareWrite      = 2
	fileAttributeNormal = 0x80
)

// IOCTL_DISK_GET_DRIVE_GEOMETRY_EX retrieves extended geometry and size.
const ioctlDiskGetDriveGeometryEx = 0x000700A0

// deviceIoControl calls DeviceIoControl from kernel32.dll.
func deviceIoControl(handle syscall.Handle, code uint32, out []byte) error {
	var returned uint32
	r1, _, e1 := syscall.NewLazyDLL("kernel32.dll").
		NewProc("DeviceIoControl").
		Call(uintptr(handle), uintptr(code), 0, 0,
			uintptr(unsafe.Pointer(&out[0])), uintptr(len(out)),
			uintptr(unsafe.Pointer(&returned)), 0)
	if r1 == 0 {
		if e1 != syscall.Errno(0) {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// openRawDevice opens a raw block device with full control + WRITE_THROUGH.
func openRawDevice(dev string) (*os.File, string) {
	return openRawDeviceFlags(dev, fileAttributeNormal|fileFlagWriteThrough)
}

// openRawDeviceFlags opens a device via CreateFileW with explicit flags.
func openRawDeviceFlags(dev string, flags uintptr) (*os.File, string) {
	desired := uintptr(genericRead | genericWrite)
	ptr, err := syscall.UTF16PtrFromString(dev)
	if err != nil {
		return nil, "invalid device path " + dev + ": " + err.Error()
	}
	k32 := syscall.NewLazyDLL("kernel32.dll")
	h, _, e := k32.NewProc("CreateFileW").Call(
		uintptr(unsafe.Pointer(ptr)), // lpFileName
		desired,                      // dwDesiredAccess
		fileShareRead|fileShareWrite, // dwShareMode
		0,                            // lpSecurityAttributes
		openExisting,                 // dwCreationDisposition
		flags,                        // dwFlagsAndAttributes
		0,                            // hTemplateFile
	)
	if h == ^uintptr(0) {
		if e == syscall.Errno(0) {
			e = syscall.EINVAL
		}
		return nil, "cannot open " + dev + ": " + e.Error()
	}
	return os.NewFile(h, dev), ""
}

// diskSize returns the total size in bytes of a block device (Windows).
func diskSize(dev string) object.Object {
	return diskSizeWindows(dev)
}

// diskDevices lists available block devices (Windows).
func diskDevices() object.Object {
	return diskDevicesWindows()
}

// diskSizeWindows returns the total size in bytes of a Windows raw device.
// Works for \\.\PhysicalDriveN (via IOCTL_DISK_GET_DRIVE_GEOMETRY_EX).
func diskSizeWindows(dev string) object.Object {
	f, e := openRawDevice(dev)
	if e != "" {
		return errObj("disk::size: " + e)
	}
	defer f.Close()

	if strings.HasPrefix(strings.ToLower(dev), `\\.\physicaldrive`) {
		buf := make([]byte, 48)
		if err := deviceIoControl(syscall.Handle(f.Fd()), ioctlDiskGetDriveGeometryEx, buf); err == nil {
			// DISK_GEOMETRY_EX: Geometry (24 bytes) then LARGE_INTEGER DiskSize at offset 24.
			size := int64(0)
			for i := 0; i < 8; i++ {
				size |= int64(buf[24+i]) << (8 * i)
			}
			if size > 0 {
				return &object.Integer{Value: size}
			}
		}
	}
	// fallback for regular files and smaller devices: seek to end.
	cur, err := f.Seek(0, os.SEEK_END)
	if err == nil && cur > 0 {
		return &object.Integer{Value: cur}
	}
	return errObj("disk::size: no size available for " + dev)
}

// diskDevicesWindows enumerates physical drives by probing \\.\PhysicalDrive0..63.
func diskDevicesWindows() object.Object {
	list := make([]object.Object, 0)
	for i := 0; i < 64; i++ {
		dev := `\\.\PhysicalDrive` + strconv.Itoa(i)
		f, err := os.OpenFile(dev, os.O_RDWR, 0)
		if err == nil {
			f.Close()
			list = append(list, &object.String{Value: dev})
		}
	}
	return &object.Array{Elements: list}
}