//go:build linux

package stdlib

import (
	"bytes"
	"os/exec"
	"strings"
	"sync"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// ---------------------------------------------------------------------------
// Linux audio backend — ALSA via `arecord` / `aplay` subprocesses.
// No cgo, no ALSA headers: keeps the final ELF a static binary. Requires the
// `alsa-utils` package (arecord, aplay) present on PATH.
//
// API (mirrors audio_windows.go so scripts are portable):
//   audio::devices()                 → Array of ALSA device lines
//   audio::record_init(r [, ch])     → INTEGER handle (starts `arecord`)
//   audio::record_take(handle)       → STRING raw S16_LE PCM (consumes buffer)
//   audio::record_stop(handle)       → NULL  (kills arecord)
//   audio::play(data [, r, ch])      → NULL  (plays via `aplay`)
// ---------------------------------------------------------------------------

var (
	recLinuxMu      sync.Mutex
	recLinuxNextID  int64 = 1
	recLinuxActive        = map[int64]*linuxRecorder{}
)

type linuxRecorder struct {
	cmd      *exec.Cmd
	rate     int64
	channels int64

	mu       sync.Mutex
	buf      bytes.Buffer
	on       bool
}

func startArecord(rate, channels int64) (*linuxRecorder, error) {
	// arecord -q -t raw -f S16_LE -r RATE -c CH -b 16 -
	cmd := exec.Command("arecord",
		"-q",
		"-t", "raw",
		"-f", "S16_LE",
		"-r", itoa(rate),
		"-c", itoa(channels),
		"-b", "16",
		"-",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	r := &linuxRecorder{cmd: cmd, rate: rate, channels: channels, on: true}
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				r.mu.Lock()
				if r.on {
					r.buf.Write(buf[:n])
				}
				r.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return r, nil
}

func stopArecord(r *linuxRecorder) {
	r.mu.Lock()
	r.on = false
	r.mu.Unlock()
	if r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
	}
	_ = r.cmd.Wait()
}

func audioLinuxRecordInit(args ...object.Object) object.Object {
	rate, ok := argsInt(args, 0)
	if !ok {
		return errObj("audio::record_init expects INTEGER samplerate [, INTEGER channels]")
	}
	channels := int64(1)
	if len(args) > 1 {
		if c, ok2 := argsInt(args, 1); ok2 {
			channels = c
		}
	}
	if rate <= 0 || channels <= 0 {
		return errObj("audio::record_init: samplerate and channels must be > 0")
	}
	r, err := startArecord(rate, channels)
	if err != nil {
		return errObj("audio::record_init: arecord: " + err.Error())
	}
	recLinuxMu.Lock()
	id := recLinuxNextID
	recLinuxNextID++
	recLinuxActive[id] = r
	recLinuxMu.Unlock()
	return &object.Integer{Value: id}
}

func audioLinuxRecordTake(args ...object.Object) object.Object {
	id, ok := argsInt(args, 0)
	if !ok {
		return errObj("audio::record_take expects INTEGER handle")
	}
	recLinuxMu.Lock()
	r, ok2 := recLinuxActive[id]
	recLinuxMu.Unlock()
	if !ok2 {
		return errObj("audio::record_take: invalid handle")
	}
	r.mu.Lock()
	data := r.buf.Bytes()
	r.buf.Reset() // "take" consumes the buffer
	r.mu.Unlock()
	if len(data) == 0 {
		return NULL
	}
	return &object.String{Value: string(data)}
}

func audioLinuxRecordStop(args ...object.Object) object.Object {
	id, ok := argsInt(args, 0)
	if !ok {
		return errObj("audio::record_stop expects INTEGER handle")
	}
	recLinuxMu.Lock()
	r, ok2 := recLinuxActive[id]
	if ok2 {
		delete(recLinuxActive, id)
	}
	recLinuxMu.Unlock()
	if !ok2 {
		return errObj("audio::record_stop: invalid handle")
	}
	stopArecord(r)
	return NULL
}

func audioLinuxPlay(args ...object.Object) object.Object {
	data := argsValue(args, 0)
	str, ok := data.(*object.String)
	if !ok {
		return errObj("audio::play expects STRING pcm data [, INTEGER samplerate, INTEGER channels]")
	}
	if len(str.Value) == 0 {
		return NULL
	}
	rate := int64(8000)
	if len(args) > 1 {
		if r, ok2 := argsInt(args, 1); ok2 {
			rate = r
		}
	}
	channels := int64(1)
	if len(args) > 2 {
		if c, ok2 := argsInt(args, 2); ok2 {
			channels = c
		}
	}
	if rate <= 0 || channels <= 0 {
		return errObj("audio::play: samplerate and channels must be > 0")
	}

	// aplay -q -t raw -f S16_LE -r RATE -c CH -b 16 -
	cmd := exec.Command("aplay",
		"-q",
		"-t", "raw",
		"-f", "S16_LE",
		"-r", itoa(rate),
		"-c", itoa(channels),
		"-b", "16",
		"-",
	)
	cmd.Stdin = strings.NewReader(str.Value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errObj("audio::play: aplay: " + err.Error() + ": " + strings.TrimSpace(string(out)))
	}
	return NULL
}

func audioLinuxDevices() []object.Object {
	var devs []object.Object
	for _, tool := range []string{"arecord", "aplay"} {
		out, err := exec.Command(tool, "-l").CombinedOutput()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "card") && strings.Contains(line, ":") {
				devs = append(devs, &object.String{Value: line})
			}
		}
	}
	return devs
}

func audioModule() *module.Module {
	m := module.New("audio")
	m.Set("devices", func(args ...object.Object) object.Object {
		return &object.Array{Elements: audioLinuxDevices()}
	})
	m.Set("record_init", audioLinuxRecordInit)
	m.Set("record_take", audioLinuxRecordTake)
	m.Set("record_stop", audioLinuxRecordStop)
	m.Set("play", audioLinuxPlay)
	m.Set("record_stop_all", func(args ...object.Object) object.Object {
		recLinuxMu.Lock()
		recLinuxActive = map[int64]*linuxRecorder{}
		recLinuxMu.Unlock()
		return NULL
	})
	m.Set("close", func(args ...object.Object) object.Object {
		return NULL
	})
	m.Desc = "Audio record/playback (ALSA: arecord/aplay)"
	return m
}
