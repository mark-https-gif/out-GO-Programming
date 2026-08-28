package stdlib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

func cryptoModule() *module.Module {
	m := module.New("crypto")
	m.Set("md5", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("crypto::md5 expects 1 argument")
		}
		s, ok := requireString("md5", args[0])
		if !ok {
			return errObj("crypto::md5 expects STRING")
		}
		h := md5.Sum([]byte(s))
		return &object.String{Value: hex.EncodeToString(h[:])}
	}).Set("sha1", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("crypto::sha1 expects 1 argument")
		}
		s, ok := requireString("sha1", args[0])
		if !ok {
			return errObj("crypto::sha1 expects STRING")
		}
		h := sha1.Sum([]byte(s))
		return &object.String{Value: hex.EncodeToString(h[:])}
	}).Set("sha256", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("crypto::sha256 expects 1 argument")
		}
		s, ok := requireString("sha256", args[0])
		if !ok {
			return errObj("crypto::sha256 expects STRING")
		}
		h := sha256.Sum256([]byte(s))
		return &object.String{Value: hex.EncodeToString(h[:])}
	}).Set("base64_encode", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("crypto::base64_encode expects 1 argument")
		}
		s, ok := requireString("base64_encode", args[0])
		if !ok {
			return errObj("crypto::base64_encode expects STRING")
		}
		return &object.String{Value: base64.StdEncoding.EncodeToString([]byte(s))}
	}).Set("base64_decode", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return errObj("crypto::base64_decode expects 1 argument")
		}
		s, ok := requireString("base64_decode", args[0])
		if !ok {
			return errObj("crypto::base64_decode expects STRING")
		}
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return errObj("crypto::base64_decode: " + err.Error())
		}
		return &object.String{Value: string(data)}
	})
	m.Desc = "Hashing and encoding (wraps Go crypto packages)"
	return m
}

func requireString(_ string, o object.Object) (string, bool) {
	s, ok := o.(*object.String)
	if !ok {
		return "", false
	}
	return s.Value, true
}

var _ = fmt.Sprintf
