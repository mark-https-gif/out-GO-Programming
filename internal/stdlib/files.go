package stdlib

import (
	"os"
	"path/filepath"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

func filesModule() *module.Module {
	m := module.New("files")
	m.Set("read", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::read expects 1 argument")
		}
		p, ok := requireString("read", args[0])
		if !ok {
			return errObj("files::read expects STRING")
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return errObj("files::read: " + err.Error())
		}
		return &object.String{Value: string(data)}
	}).Set("write", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("files::write expects 2 arguments")
		}
		p, ok1 := requireString("write", args[0])
		c, ok2 := requireString("write", args[1])
		if !ok1 || !ok2 {
			return errObj("files::write expects STRING, STRING")
		}
		if err := os.WriteFile(p, []byte(c), 0644); err != nil {
			return errObj("files::write: " + err.Error())
		}
		return NULL
	}).Set("append", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return errObj("files::append expects 2 arguments")
		}
		p, ok1 := requireString("append", args[0])
		c, ok2 := requireString("append", args[1])
		if !ok1 || !ok2 {
			return errObj("files::append expects STRING, STRING")
		}
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errObj("files::append: " + err.Error())
		}
		defer f.Close()
		if _, err := f.WriteString(c + "\n"); err != nil {
			return errObj("files::append: " + err.Error())
		}
		return NULL
	}).Set("exists", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::exists expects 1 argument")
		}
		p, ok := requireString("exists", args[0])
		if !ok {
			return errObj("files::exists expects STRING")
		}
		_, err := os.Stat(p)
		return &object.Boolean{Value: err == nil}
	}).Set("list", func(args ...object.Object) object.Object {
		dir := "."
		if len(args) == 1 {
			if s, ok := requireString("list", args[0]); ok {
				dir = s
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return errObj("files::list: " + err.Error())
		}
		elems := make([]object.Object, len(entries))
		for i, e := range entries {
			if e.IsDir() {
				elems[i] = &object.String{Value: e.Name() + "/"}
			} else {
				elems[i] = &object.String{Value: e.Name()}
			}
		}
		return &object.Array{Elements: elems}
	}).Set("mkdir", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::mkdir expects 1 argument")
		}
		p, ok := requireString("mkdir", args[0])
		if !ok {
			return errObj("files::mkdir expects STRING")
		}
		if err := os.MkdirAll(p, 0755); err != nil {
			return errObj("files::mkdir: " + err.Error())
		}
		return NULL
	}).Set("remove", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::remove expects 1 argument")
		}
		p, ok := requireString("remove", args[0])
		if !ok {
			return errObj("files::remove expects STRING")
		}
		if err := os.RemoveAll(p); err != nil {
			return errObj("files::remove: " + err.Error())
		}
		return NULL
	}).Set("basename", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::basename expects 1 argument")
		}
		p, ok := requireString("basename", args[0])
		if !ok {
			return errObj("files::basename expects STRING")
		}
		return &object.String{Value: filepath.Base(p)}
	}).Set("ext", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("files::ext expects 1 argument")
		}
		p, ok := requireString("ext", args[0])
		if !ok {
			return errObj("files::ext expects STRING")
		}
		return &object.String{Value: filepath.Ext(p)}
	})
	m.Desc = "File system operations (wraps Go os/path/filepath)"
	return m
}
