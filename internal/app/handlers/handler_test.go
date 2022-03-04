package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
)

func ExampleHandler_Save() {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	// Allocation handler and storage
	h := New(zap.NewNop(), rep)

	r = strings.NewReader("http://newlink.ru")
	request := httptest.NewRequest(http.MethodPost, "/", r)

	rtr.HandleFunc("/", h.Save)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}

func TestHandler_Save(t *testing.T) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	// Allocation handler and storage
	h := New(zap.NewNop(), rep)

	r = strings.NewReader("http://newlink.ru")
	request := httptest.NewRequest(http.MethodPost, "/", r)

	rtr.HandleFunc("/", h.Save)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusCreated, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()

	assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))

}

func ExampleHandler_Get() {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}
	h := New(zap.NewNop(), rep)

	r = strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/GMWJGSAPGA", r)

	rtr.HandleFunc("/{id:.+}", h.Get)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}

func TestHandler_Get(t *testing.T) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	// Allocation handler and storage
	h := New(zap.NewNop(), rep)

	r = strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/xxx", r)

	rtr.HandleFunc("/{id:.+}", h.Get)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()

	assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))
}

func ExampleHandler_SaveJSON() {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)

	r = strings.NewReader("{\"unknown\": \"http://vtest.com\"}")
	request := httptest.NewRequest(http.MethodPost, "/", r)

	rtr.HandleFunc("/", h.SaveJSON)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}

func TestHandler_SaveJSON(t *testing.T) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)

	r = strings.NewReader("{\"urfxxxx\": \"http://vtest.com\"}")
	request := httptest.NewRequest(http.MethodPost, "/", r)

	rtr.HandleFunc("/", h.SaveJSON)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()

	assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	readLine := strings.TrimSuffix(string(resBody), "\n")

	assert.Equal(t, "unknown url", readLine)
}

func ExampleHandler_GetUrls() {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)
	r = strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

	rtr.HandleFunc("/user/urls", h.GetUrls)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}

func TestHandler_GetUrls(t *testing.T) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)

	r = strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

	rtr.HandleFunc("/user/urls", h.GetUrls)
	rtr.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusNoContent, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()

	assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))

}

func ExampleNew() {
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)

	fmt.Println(h)
}

func TestNew(t *testing.T) {
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	h := New(zap.NewNop(), rep)

	assert.IsType(t, &Handler{}, h)
}
