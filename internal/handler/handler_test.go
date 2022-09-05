package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHeaders(t *testing.T) {
	t.Parallel()

	origin := "localhost"
	body := "hello world"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Origin", origin)

	w := httptest.NewRecorder()

	CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("bad response: %s (%d)", res.Status, res.StatusCode)
	}

	if v := res.Header.Get(AllowMethodsKey); v != AllowedMethods {
		t.Errorf("want: %s, got: %s", AllowedMethods, v)
	}

	if v := res.Header.Get(AllowOriginKey); v != origin {
		t.Errorf("want: %s, got: %s", origin, v)
	}

	if v := res.Header.Get(AllowHeadersKey); v != AllowedHeaders {
		t.Errorf("want: %s, got: %s", AllowedHeaders, v)
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != body {
		t.Errorf("want: %s, got: %s", body, string(contents))
	}

	t.Logf("full response: %+v", res)
}

func TestCORSPreflight(t *testing.T) {
	t.Parallel()

	origin := "localhost"
	body := "hello world"

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Add("Origin", origin)

	w := httptest.NewRecorder()

	CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, body)
	}))(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("bad response: %s (%d)", res.Status, res.StatusCode)
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != "" {
		t.Errorf("wanted nothing, got: %s", string(contents))
	}

	t.Logf("full response: %+v", res)
}
