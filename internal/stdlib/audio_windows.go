//go:build windows

package stdlib

import (
	"encoding/binary"
	"sync"
	"syscall"
	"unsafe"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// ---------------------------------------------------------------------------
// winmm.dll / kernel32.dll bindings (pure syscall — no CGO)
// ---------------------------------------------------------------------------

var (
	winmm    = syscall.NewLazyDLL("winmm.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pWaveInGetNumDevs    = winmm.NewProc("waveInGetNumDevs")
	pWaveInGetDevCapsW   = winmm.NewProc("waveInGetDevCapsW")
	pWaveInOpen          = winmm.NewProc("waveInOpen")
	pWaveInClose         = winmm.NewProc("waveInClose")
	pWaveInStart         = winmm.NewProc("waveInStart")
	pWaveInStop          = winmm.NewProc("waveInStop")
	pWaveInPrepareHeader = winmm.NewProc("waveInPrepareHeader")
	pWaveInUnprepareHeader = winmm.NewProc("waveInUnprepareHeader")
	pWaveInAddBuffer     = winmm.NewProc("waveInAddBuffer")

	pWaveOutGetNumDevs     = winmm.NewProc("waveOutGetNumDevs")
	pWaveOutGetDevCapsW    = winmm.NewProc("waveOutGetDevCapsW")
	pWaveOutOpen           = winmm.NewProc("waveOutOpen")
	pWaveOutClose          = winmm.NewProc("waveOutClose")
	pWaveOutWrite          = winmm.NewProc("waveOutWrite")
	pWaveOutPrepareHeader  = winmm.NewProc("waveOutPrepareHeader")
	pWaveOutUnprepareHeader = winmm.NewProc("waveOutUnprepareHeader")

	pCreateEvent         = kernel32.NewProc("CreateEventW")
	pWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	pResetEvent          = kernel32.NewProc("ResetEvent")
	pCloseHandle         = kernel32.NewProc("CloseHandle")
)

const (
	waveFormatPCM = 1
	waveMapper    = 0xFFFFFFFF
	whdrDone      = 0x00000001
	callBackEvent = 0x00050000

	numRecBufs = 4
	recBufMs   = 100 // ms of audio per buffer
)

// WAVEHDR layout differs between x86 and x64; compute offsets at init time.
var (
	ptrSize    uintptr
	hdrSz      uintptr
	hdrBytesOf uintptr
	hdrFlagsOf uintptr
)

func init() {
	ptrSize = unsafe.Sizeof(uintptr(0))
	if ptrSize == 8 { // x64
		hdrSz = 48
		hdrBytesOf = 12
		hdrFlagsOf = 24
	} else { // x86
		hdrSz = 32
		hdrBytesOf = 8
		hdrFlagsOf = 16
	}
}

// ---------------------------------------------------------------------------
// Struct builders (little-endian byte slices used as pointers to winmm)
// ---------------------------------------------------------------------------

// makeWfx builds a WAVEFORMATEX (18 bytes).
func makeWfx(rate, channels, bits int) []byte {
	b := make([]byte, 18)
	ba := uint16(channels * bits / 8)
	binary.LittleEndian.PutUint16(b[0:], waveFormatPCM)
	binary.LittleEndian.PutUint16(b[2:], uint16(channels))
	binary.LittleEndian.PutUint32(b[4:], uint32(rate))
	binary.LittleEndian.PutUint32(b[8:], uint32(rate)*uint32(ba))
	binary.LittleEndian.PutUint16(b[12:], ba)
	binary.LittleEndian.PutUint16(b[14:], uint16(bits))
	binary.LittleEndian.PutUint16(b[16:], 0)
	return b
}

// makeHdr builds a WAVEHDR pointing at data.
func makeHdr(data []byte) []byte {
	h := make([]byte, hdrSz)
	if ptrSize == 8 {
		binary.LittleEndian.PutUint64(h[0:], uint64(uintptr(unsafe.Pointer(&data[0]))))
	} else {
		binary.LittleEndian.PutUint32(h[0:], uint32(uintptr(unsafe.Pointer(&data[0]))))
	}
	binary.LittleEndian.PutUint32(h[8:], uint32(len(data)))
	return h
}

func hdrFlags(h []byte) uint32   { return binary.LittleEndian.Uint32(h[hdrFlagsOf:]) }
func setHdrFlags(h []byte, f uint32) { binary.LittleEndian.PutUint32(h[hdrFlagsOf:], f) }
func hdrBytesRecorded(h []byte) uint32 { return binary.LittleEndian.Uint32(h[hdrBytesOf:]) }

// ---------------------------------------------------------------------------
// Device enumeration
// ---------------------------------------------------------------------------

type waveInCaps struct {
	wMid       uint32
	wPid       uint32
	vDriverVer uint32
	szPname    [32]uint16
	dwFormats  uint32
	wChannels  uint16
	wReserved1 uint16
}

type waveOutCaps struct {
	wMid       uint32
	wPid       uint32
	vDriverVer uint32
	szPname    [32]uint16
	dwFormats  uint32
	wChannels  uint16
	wReserved1 uint16
	dwSupport  uint32
}

func enumerateWaveIn() []object.Object {
	n, _, _ := pWaveInGetNumDevs.Call()
	var devs []object.Object
	for i := uintptr(0); i < n; i++ {
		var caps waveInCaps
		pWaveInGetDevCapsW.Call(i, uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps))
		devs = append(devs, &object.String{Value: syscall.UTF16ToString(caps.szPname[:])})
	}
	return devs
}

func enumerateWaveOut() []object.Object {
	n, _, _ := pWaveOutGetNumDevs.Call()
	var devs []object.Object
	for i := uintptr(0); i < n; i++ {
		var caps waveOutCaps
		pWaveOutGetDevCapsW.Call(i, uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps))
		devs = append(devs, &object.String{Value: syscall.UTF16ToString(caps.szPname[:])})
	}
	return devs
}

// ---------------------------------------------------------------------------
// Recording
// ---------------------------------------------------------------------------

type recorder struct {
	handle   uintptr
	event    uintptr
	rate     int
	channels int

	mu      sync.Mutex
	buf     []byte
	running bool
	done    chan struct{}

	dataBufs [numRecBufs][]byte
	hdrBufs  [numRecBufs][]byte
}

var (
	recMu     sync.Mutex
	recorders = make(map[int64]*recorder)
	recNextID int64 = 1
)

func (r *recorder) recordLoop() {
	defer close(r.done)
	for {
		// Wait for the event (200ms timeout so we can notice r.running == false).
		ret, _, _ := pWaitForSingleObject.Call(r.event, 200)
		pResetEvent.Call(r.event)

		r.mu.Lock()
		if !r.running {
			r.mu.Unlock()
			return
		}
		for i := 0; i < numRecBufs; i++ {
			if hdrFlags(r.hdrBufs[i])&whdrDone == 0 {
				continue
			}
			if n := hdrBytesRecorded(r.hdrBufs[i]); n > 0 {
				chunk := make([]byte, n)
				copy(chunk, r.dataBufs[i][:n])
				r.buf = append(r.buf, chunk...)
			}
			setHdrFlags(r.hdrBufs[i], 0)
			pWaveInAddBuffer.Call(r.handle, uintptr(unsafe.Pointer(&r.hdrBufs[i][0])), hdrSz)
		}
		r.mu.Unlock()
		_ = ret
	}
}

func createEvent() (uintptr, bool) {
	h, _, _ := pCreateEvent.Call(0, 1, 0, 0) // manualReset=1, initialState=1
	return h, h != 0
}

func audioRecordInit(args []object.Object) object.Object {
	rate, ok1 := argsInt(args, 0)
	if !ok1 {
		return errObj("audio::record_init expects INTEGER samplerate [, INTEGER channels]")
	}
	channels := int64(1)
	if len(args) > 1 {
		if ch, ok := argsInt(args, 1); ok {
			channels = ch
		}
	}
	deviceID := int64(-1) // -1 = WAVE_MAPPER
	if len(args) > 2 {
		if d, ok := argsInt(args, 2); ok {
			deviceID = d
		}
	}
	if rate <= 0 || channels <= 0 {
		return errObj("audio::record_init: samplerate and channels must be > 0")
	}

	ev, ok := createEvent()
	if !ok {
		return errObj("audio::record_init: CreateEvent failed")
	}

	wfx := makeWfx(int(rate), int(channels), 16)
	bufBytes := int(rate) * int(channels) * 2 / 1000 * recBufMs

	devParam := ^uintptr(0) // WAVE_MAPPER = (UINT)-1
	if deviceID >= 0 {
		devParam = uintptr(deviceID)
	}

	var hdev uintptr
	ret, _, e := pWaveInOpen.Call(
		uintptr(unsafe.Pointer(&hdev)),
		devParam,
		uintptr(unsafe.Pointer(&wfx[0])),
		ev,
		0,
		callBackEvent,
	)
	if ret != 0 && deviceID < 0 {
		// WAVE_MAPPER can fail on some drivers; fall back to the first
		// input device that opens.
		n, _, _ := pWaveInGetNumDevs.Call()
		for i := uintptr(0); i < n; i++ {
			ret, _, e = pWaveInOpen.Call(
				uintptr(unsafe.Pointer(&hdev)),
				i,
				uintptr(unsafe.Pointer(&wfx[0])),
				ev,
				0,
				callBackEvent,
			)
			if ret == 0 {
				break
			}
		}
	}
	if ret != 0 {
		pCloseHandle.Call(ev)
		return errObj("audio::record_init: waveInOpen failed (MME " + itoa(int64(ret)) + ") " + e.Error())
	}

	r := &recorder{
		handle:   hdev,
		event:    ev,
		rate:     int(rate),
		channels: int(channels),
		running:  true,
		done:     make(chan struct{}),
	}

	for i := 0; i < numRecBufs; i++ {
		r.dataBufs[i] = make([]byte, bufBytes)
		r.hdrBufs[i] = makeHdr(r.dataBufs[i])
		if ret, _, _ := pWaveInPrepareHeader.Call(hdev, uintptr(unsafe.Pointer(&r.hdrBufs[i][0])), hdrSz); ret != 0 {
			pWaveInClose.Call(hdev)
			pCloseHandle.Call(ev)
			return errObj("audio::record_init: waveInPrepareHeader failed")
		}
		if ret, _, _ := pWaveInAddBuffer.Call(hdev, uintptr(unsafe.Pointer(&r.hdrBufs[i][0])), hdrSz); ret != 0 {
			pWaveInClose.Call(hdev)
			pCloseHandle.Call(ev)
			return errObj("audio::record_init: waveInAddBuffer failed")
		}
	}

	if ret, _, _ := pWaveInStart.Call(hdev); ret != 0 {
		pWaveInClose.Call(hdev)
		pCloseHandle.Call(ev)
		return errObj("audio::record_init: waveInStart failed")
	}

	go r.recordLoop()

	recMu.Lock()
	id := recNextID
	recNextID++
	recorders[id] = r
	recMu.Unlock()
	return &object.Integer{Value: id}
}

func audioRecordTake(args []object.Object) object.Object {
	id, ok := argsInt(args, 0)
	if !ok {
		return errObj("audio::record_take expects INTEGER handle")
	}
	recMu.Lock()
	r, ok := recorders[id]
	recMu.Unlock()
	if !ok {
		return errObj("audio::record_take: invalid handle")
	}
	r.mu.Lock()
	data := r.buf
	r.buf = nil
	r.mu.Unlock()
	if len(data) == 0 {
		return NULL
	}
	return &object.String{Value: string(data)}
}

func audioRecordStop(args []object.Object) object.Object {
	id, ok := argsInt(args, 0)
	if !ok {
		return errObj("audio::record_stop expects INTEGER handle")
	}
	recMu.Lock()
	r, ok := recorders[id]
	if ok {
		delete(recorders, id)
	}
	recMu.Unlock()
	if !ok {
		return errObj("audio::record_stop: invalid handle")
	}

	r.mu.Lock()
	r.running = false
	r.mu.Unlock()

	pWaveInStop.Call(r.handle)
	pWaveInClose.Call(r.handle)
	<-r.done
	pCloseHandle.Call(r.event)
	return NULL
}

// ---------------------------------------------------------------------------
// Playback
// ---------------------------------------------------------------------------

func audioPlay(args []object.Object) object.Object {
	data, ok1 := requireString("play", argsValue(args, 0))
	if !ok1 {
		return errObj("audio::play expects STRING pcm data [, INTEGER samplerate, INTEGER channels]")
	}
	if len(data) == 0 {
		return NULL
	}

	rate, channels := 8000, 1
	if len(args) > 1 {
		if r, ok := argsInt(args, 1); ok {
			rate = int(r)
		}
	}
	if len(args) > 2 {
		if ch, ok := argsInt(args, 2); ok {
			channels = int(ch)
		}
	}
	if rate <= 0 || channels <= 0 {
		return errObj("audio::play: samplerate and channels must be > 0")
	}

	ev, ok := createEvent()
	if !ok {
		return errObj("audio::play: CreateEvent failed")
	}
	defer pCloseHandle.Call(ev)

	wfx := makeWfx(rate, channels, 16)
	var hdev uintptr
	ret, _, e := pWaveOutOpen.Call(
		uintptr(unsafe.Pointer(&hdev)),
		^uintptr(0), // WAVE_MAPPER
		uintptr(unsafe.Pointer(&wfx[0])),
		ev,
		0,
		callBackEvent,
	)
	if ret != 0 {
		n, _, _ := pWaveOutGetNumDevs.Call()
		for i := uintptr(0); i < n; i++ {
			ret, _, e = pWaveOutOpen.Call(
				uintptr(unsafe.Pointer(&hdev)),
				i,
				uintptr(unsafe.Pointer(&wfx[0])),
				ev,
				0,
				callBackEvent,
			)
			if ret == 0 {
				break
			}
		}
	}
	if ret != 0 {
		return errObj("audio::play: waveOutOpen failed (MME " + itoa(int64(ret)) + ") " + e.Error())
	}
	defer pWaveOutClose.Call(hdev)

	pcm := make([]byte, len(data))
	copy(pcm, data)
	hdr := makeHdr(pcm)

	if ret, _, _ := pWaveOutPrepareHeader.Call(hdev, uintptr(unsafe.Pointer(&hdr[0])), hdrSz); ret != 0 {
		return errObj("audio::play: waveOutPrepareHeader failed")
	}
	if ret, _, _ := pWaveOutWrite.Call(hdev, uintptr(unsafe.Pointer(&hdr[0])), hdrSz); ret != 0 {
		pWaveOutUnprepareHeader.Call(hdev, uintptr(unsafe.Pointer(&hdr[0])), hdrSz)
		return errObj("audio::play: waveOutWrite failed")
	}

	pWaitForSingleObject.Call(ev, 10000) // wait up to 10s for playback
	pWaveOutUnprepareHeader.Call(hdev, uintptr(unsafe.Pointer(&hdr[0])), hdrSz)
	return NULL
}

// ---------------------------------------------------------------------------
// Module registration
// ---------------------------------------------------------------------------

func audioModule() *module.Module {
	m := module.New("audio")

	// devices() → array of input device names followed by output device names
	m.Set("devices", func(args ...object.Object) object.Object {
		all := enumerateWaveIn()
		all = append(all, enumerateWaveOut()...)
		return &object.Array{Elements: all}
	})

	// record_init(samplerate [, channels]) → handle
	m.Set("record_init", func(args ...object.Object) object.Object {
		return audioRecordInit(args)
	})

	// record_take(handle) → STRING of raw 16-bit PCM bytes, or null if none yet
	m.Set("record_take", func(args ...object.Object) object.Object {
		return audioRecordTake(args)
	})

	// record_stop(handle) → NULL
	m.Set("record_stop", func(args ...object.Object) object.Object {
		return audioRecordStop(args)
	})

	// play(data [, samplerate, channels]) → NULL (blocks until playback done)
	m.Set("play", func(args ...object.Object) object.Object {
		return audioPlay(args)
	})

	// close — no-op (use record_stop for recorders)
	m.Set("close", func(args ...object.Object) object.Object {
		return NULL
	})

	m.Desc = "Audio capture/playback via Windows waveIn/waveOut (PCM 16-bit)"
	return m
}
