//go:build !windows && !linux

package stdlib

import (
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

func audioModule() *module.Module {
	m := module.New("audio")
	m.Set("devices", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Set("record_init", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Set("record_stop", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Set("record_take", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Set("play", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Set("close", func(args ...object.Object) object.Object {
		return errObj("audio:: not supported on this platform")
	})
	m.Desc = "Audio record/playback (Windows waveIn/waveOut only)"
	return m
}
