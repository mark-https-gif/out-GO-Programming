//go:build !windows

package stdlib

import (
	"os"
	"strings"

	"github.com/out-lang/out/internal/object"
)

// openRawDevice opens a raw block device on unix (usually needs root).
func openRawDevice(dev string) (*os.File, string) {
	f, err := os.OpenFile(dev, os.O_RDWR, 0)
	if err != nil {
		return nil, err.Error()
	}
	return f, ""
}

// diskSize returns the total size in bytes of a block device (unix).
func diskSize(dev string) object.Object {
	return diskSizeUnix(dev)
}

// diskDevices lists available block devices (unix).
func diskDevices() object.Object {
	return diskDevicesUnix()
}

// diskSizeUnix returns the size of a block device on unix systems.
func diskSizeUnix(dev string) object.Object {
	f, e := openRawDevice(dev)
	if e != "" {
		return errObj("disk::size: " + e)
	}
	defer f.Close()
	cur, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return errObj("disk::size: " + err.Error())
	}
	// SEEK_END on block devices requires opening with the correct mode;
	// if size is invalid fall back to -1 and let the caller decide.
	return &object.Integer{Value: cur}
}

// diskDevicesUnix lists block devices under /dev on unix systems.
func diskDevicesUnix() object.Object {
	list := make([]object.Object, 0)
	add := func(prefix string) {
		entries, err := os.ReadDir("/dev")
		if err != nil {
			return
		}
		for _, e := range entries {
			name := e.Name()
			if !isBlockDeviceCandidate(prefix, name) {
				continue
			}
			path := "/dev/" + name
			if fi, err := os.Stat(path); err == nil && fi.Mode()&os.ModeDevice != 0 {
				list = append(list, &object.String{Value: path})
			}
		}
	}
	add("sd")
	add("nvme")
	add("hd")
	return &object.Array{Elements: list}
}

// isBlockDeviceCandidate filters out partitions (sdX1, nvmeXnYpZ, hdX1).
func isBlockDeviceCandidate(prefix, name string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	rest := strings.TrimPrefix(name, prefix)
	for _, r := range rest {
		if r >= '0' && r <= '9' {
			return false
		}
	}
	return true
}