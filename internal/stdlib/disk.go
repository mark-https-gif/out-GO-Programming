package stdlib

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// Handles returned by disk::open map to real *os.File kept in this registry.
var (
	diskMu      sync.Mutex
	diskHandles = make(map[int64]*os.File)
	diskNextID  int64 = 1
)

// diskModule exposes raw block-device / USB access.
// On Windows: \\.\PhysicalDriveN or \\.\X:
// On Linux: /dev/sdX, /dev/nvme*, /dev/hd*
//
// Two styles:
//   - one-shot helpers: write_raw, write_file, write_file_at, write_at, read_at
//   - handle API (recommended for long writes):
//     h = disk::open("\\.\F:", "write"); disk::write(h, data); disk::close(h)
//
// Handle API opens devices with FILE_FLAG_WRITE_THROUGH on Windows so data is
// flushed to the hardware immediately. Writes are sector-aligned (read-modify-write
// for a trailing partial sector) so no neighbouring data is corrupted.
func diskModule() *module.Module {
	m := module.New("disk")

	m.Set("devices", func(args ...object.Object) object.Object {
		return diskDevices()
	})

	m.Set("size", func(args ...object.Object) object.Object {
		dev, ok := requireString("size", argsValue(args, 0))
		if !ok {
			return errObj("disk::size expects STRING device path")
		}
		return diskSize(dev)
	})

	m.Set("open", func(args ...object.Object) object.Object {
		dev, ok := requireString("open", argsValue(args, 0))
		if !ok {
			return errObj("disk::open expects STRING device path")
		}
		readonly := false
		if len(args) > 1 {
			mode, ok := args[1].(*object.String)
			if !ok {
				return errObj("disk::open mode must be STRING ('read' or 'write')")
			}
			switch mode.Value {
			case "read":
				readonly = true
			case "write", "rw":
			default:
				return errObj("disk::open mode must be 'read' or 'write'")
			}
		}
		return diskOpen(dev, readonly)
	})

	m.Set("close", func(args ...object.Object) object.Object {
		h, ok := argsInt(args, 0)
		if !ok {
			return errObj("disk::close expects INTEGER handle")
		}
		return diskClose(h)
	})

	m.Set("write", func(args ...object.Object) object.Object {
		h, ok1 := argsInt(args, 0)
		data, ok2 := requireString("write", argsValue(args, 1))
		if !ok1 || !ok2 {
			return errObj("disk::write expects INTEGER handle, STRING bytes")
		}
		return diskHandleWrite(h, data)
	})

	m.Set("read", func(args ...object.Object) object.Object {
		h, ok1 := argsInt(args, 0)
		n, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("disk::read expects INTEGER handle, INTEGER length")
		}
		if n < 0 {
			return errObj("disk::read: length must be >= 0")
		}
		return diskHandleRead(h, n)
	})

	m.Set("seek", func(args ...object.Object) object.Object {
		h, ok1 := argsInt(args, 0)
		off, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("disk::seek expects INTEGER handle, INTEGER offset")
		}
		return diskHandleSeek(h, off)
	})

	m.Set("flush", func(args ...object.Object) object.Object {
		h, ok := argsInt(args, 0)
		if !ok {
			return errObj("disk::flush expects INTEGER handle")
		}
		return diskHandleFlush(h)
	})

	m.Set("write_raw", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("disk::write_raw expects 2 arguments (device, bytes)")
		}
		dev, ok1 := requireString("write_raw", args[0])
		content, ok2 := requireString("write_raw", args[1])
		if !ok1 || !ok2 {
			return errObj("disk::write_raw expects STRING device, STRING bytes")
		}
		return diskWrite(dev, []byte(content), 0, true)
	})

	m.Set("write_file", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("disk::write_file expects 2 arguments (device, path)")
		}
		dev, ok1 := requireString("write_file", args[0])
		path, ok2 := requireString("write_file", args[1])
		if !ok1 || !ok2 {
			return errObj("disk::write_file expects STRING device, STRING path")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return errObj("disk::write_file: " + err.Error())
		}
		return diskWrite(dev, data, 0, true)
	})

	m.Set("write_file_at", func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return errObj("disk::write_file_at expects 3 arguments (device, path, offset)")
		}
		dev, ok1 := requireString("write_file_at", args[0])
		path, ok2 := requireString("write_file_at", args[1])
		off, ok3 := argsInt(args, 2)
		if !ok1 || !ok2 || !ok3 {
			return errObj("disk::write_file_at expects STRING device, STRING path, INTEGER offset")
		}
		if off < 0 {
			return errObj("disk::write_file_at: offset must be >= 0")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return errObj("disk::write_file_at: " + err.Error())
		}
		return diskWrite(dev, data, off, true)
	})

	m.Set("write_at", func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return errObj("disk::write_at expects 3 arguments (device, offset, bytes)")
		}
		dev, ok1 := requireString("write_at", args[0])
		off, ok2 := argsInt(args, 1)
		content, ok3 := requireString("write_at", args[2])
		if !ok1 || !ok2 || !ok3 {
			return errObj("disk::write_at expects STRING device, INTEGER offset, STRING bytes")
		}
		if off < 0 {
			return errObj("disk::write_at: offset must be >= 0")
		}
		return diskWrite(dev, []byte(content), off, true)
	})

	m.Set("read_at", func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return errObj("disk::read_at expects 3 arguments (device, offset, length)")
		}
		dev, ok1 := requireString("read_at", args[0])
		off, ok2 := argsInt(args, 1)
		n, ok3 := argsInt(args, 2)
		if !ok1 || !ok2 || !ok3 {
			return errObj("disk::read_at expects STRING device, INTEGER offset, INTEGER length")
		}
		if off < 0 || n < 0 {
			return errObj("disk::read_at: offset and length must be >= 0")
		}
		data, err := diskRead(dev, off, n)
		if err != "" {
			return errObj("disk::read_at: " + err)
		}
		return &object.String{Value: string(data)}
	})

	m.Desc = "Raw block-device / USB access (\\\\.\\PhysicalDriveN, \\\\.\\X:, /dev/sdX)"
	return m
}

func argsValue(args []object.Object, i int) object.Object {
	if i < len(args) {
		return args[i]
	}
	return NULL
}

func argsInt(args []object.Object, i int) (int64, bool) {
	if i >= len(args) {
		return 0, false
	}
	v, ok := args[i].(*object.Integer)
	if !ok {
		return 0, false
	}
	return v.Value, true
}

// diskOpen opens a raw device and returns an integer handle for later calls.
// On Windows the file is opened with FILE_FLAG_WRITE_THROUGH.
func diskOpen(dev string, readonly bool) object.Object {
	f, e := openRawDevice(dev)
	if e != "" {
		if readonly {
			ff, err := os.OpenFile(dev, os.O_RDONLY, 0)
			if err == nil {
				f = ff
				e = ""
			}
		}
		if e != "" {
			return errObj("disk::open: " + e)
		}
	}

	diskMu.Lock()
	id := diskNextID
	diskNextID++
	diskHandles[id] = f
	diskMu.Unlock()
	return &object.Integer{Value: id}
}

// diskClose closes a handle returned by disk::open.
func diskClose(h int64) object.Object {
	diskMu.Lock()
	f, ok := diskHandles[h]
	if ok {
		delete(diskHandles, h)
	}
	diskMu.Unlock()
	if !ok {
		return errObj("disk::close: invalid handle")
	}
	if err := f.Close(); err != nil {
		return errObj("disk::close: " + err.Error())
	}
	return NULL
}

// diskGetHandle returns the *os.File behind a handle id.
func diskGetHandle(h int64) (*os.File, string) {
	diskMu.Lock()
	defer diskMu.Unlock()
	f, ok := diskHandles[h]
	if !ok {
		return nil, "invalid handle"
	}
	return f, ""
}

// writeSectorAligned writes data to f at its current position, padding the
// trailing partial sector with zeros. Sector alignment matters for raw disks.
func writeSectorAligned(f *os.File, data []byte) error {
	const sector = 512

	if len(data)%sector == 0 {
		_, err := f.Write(data)
		return err
	}

	// pad tail to a whole sector without restoring previous content
	rem := len(data) % sector
	data = append(data, make([]byte, sector-rem)...)
	_, err := f.Write(data)
	return err
}

// diskHandleWrite appends a data chunk to a handle at its current position.
func diskHandleWrite(h int64, data string) object.Object {
	f, e := diskGetHandle(h)
	if e != "" {
		return errObj("disk::write: " + e)
	}
	if err := writeSectorAligned(f, []byte(data)); err != nil {
		return errObj("disk::write: " + err.Error())
	}
	return NULL
}

// diskHandleRead reads n bytes from a handle at its current position.
func diskHandleRead(h int64, n int64) object.Object {
	f, e := diskGetHandle(h)
	if e != "" {
		return errObj("disk::read: " + e)
	}
	buf := make([]byte, n)
	got, err := f.Read(buf)
	if err != nil && got == 0 {
		return errObj("disk::read: " + err.Error())
	}
	return &object.String{Value: string(buf[:got])}
}

// diskHandleSeek moves the handle position to a byte offset (from start).
func diskHandleSeek(h int64, offset int64) object.Object {
	f, e := diskGetHandle(h)
	if e != "" {
		return errObj("disk::seek: " + e)
	}
	if offset < 0 {
		return errObj("disk::seek: offset must be >= 0")
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return errObj("disk::seek: " + err.Error())
	}
	return NULL
}

// diskHandleFlush flushes buffered data to the hardware (FlushFileBuffers/Sync).
func diskHandleFlush(h int64) object.Object {
	f, e := diskGetHandle(h)
	if e != "" {
		return errObj("disk::flush: " + e)
	}
	if err := f.Sync(); err != nil {
		return errObj("disk::flush: " + err.Error())
	}
	return NULL
}

// diskWrite writes data to a raw device at a byte offset.
// When padSector is true the tail is zero-padded to a whole 512-byte sector.
func diskWrite(dev string, data []byte, offset int64, padSector bool) object.Object {
	f, e := openRawDevice(dev)
	if e != "" {
		return errObj("disk::write: " + e)
	}
	defer f.Close()

	if padSector {
		if rem := len(data) % 512; rem != 0 {
			data = append(data, make([]byte, 512-rem)...)
		}
	}

	written := 0
	for written < len(data) {
		n, err := f.WriteAt(data[written:], offset+int64(written))
		if err != nil {
			return errObj("disk::write: error at offset " + strconv.FormatInt(offset+int64(written), 10) + ": " + err.Error())
		}
		written += n
	}
	f.Sync()
	return NULL
}

// diskRead reads n bytes from a device starting at a byte offset.
func diskRead(dev string, offset, n int64) ([]byte, string) {
	f, e := openRawDevice(dev)
	if e != "" {
		return nil, e
	}
	defer f.Close()
	buf := make([]byte, n)
	got, err := f.ReadAt(buf, offset)
	if err != nil && got == 0 {
		return nil, "read error at offset " + strconv.FormatInt(offset, 10) + ": " + err.Error()
	}
	return buf[:got], ""
}

// diskSizesStr is used by platform implementations for friendly number output.
func diskSizesStr(n int64) string {
	return fmt.Sprintf("%d bytes (%.2f MB)", n, float64(n)/(1024*1024))
}
