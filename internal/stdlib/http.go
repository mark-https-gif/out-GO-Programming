package stdlib

import (
	"io"
	"net/http"
	"time"

	"github.com/out-lang/out/internal/object"
)

// httpGet performs a GET request and returns a friendly OUT structure.
func httpGet(url string) object.Object {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return errObj("http::get failed: " + err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errObj("http::get read failed: " + err.Error())
	}

	ok := false
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ok = true
	}

	status := &object.Integer{Value: int64(resp.StatusCode)}
	bodyStr := &object.String{Value: string(body)}
	okBool := &object.Boolean{Value: ok}

	pairs := make(map[uint64]object.HashPair)
	add := func(k *object.String, v object.Object) {
		pairs[object.HashKey(k)] = object.HashPair{Key: k, Value: v}
	}
	add(&object.String{Value: "status"}, status)
	add(&object.String{Value: "body"}, bodyStr)
	add(&object.String{Value: "ok"}, okBool)

	return &object.Hash{Pairs: pairs}
}

var _ = httpResult{}
