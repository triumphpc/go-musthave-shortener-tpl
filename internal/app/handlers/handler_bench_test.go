// Create handler Save load
//
// go test -bench=BenchmarkHandler_Save -benchmem -benchtime=10000x -memprofile base.pprof
// go tool pprof -http=":9090" bench.test base.pprof

package handlers

import (
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkHandler_Save(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	c := configs.Instance()
	// Allocation handler and storage
	h := New(c.Logger, c.Storage)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "http://test_link_" + strconv.Itoa(i) + ".ru"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer() //
		rtr.HandleFunc("/", h.Save)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()
		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

//
func BenchmarkHandler_Get(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	c := configs.Instance()
	// Allocation handler and storage
	h := New(c.Logger, c.Storage)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "/AzLn727sq" + strconv.Itoa(i)
		request := httptest.NewRequest(http.MethodGet, st, r)

		b.StartTimer() //
		rtr.HandleFunc("/{id:.+}", h.Get)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

//
func BenchmarkHandler_GetUrls(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	c := configs.Instance()
	// Allocation handler and storage
	h := New(c.Logger, c.Storage)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers

		request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

		b.StartTimer() //
		rtr.HandleFunc("/user/urls", h.Get)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		res.Body.Close()
	}
}

//
func BenchmarkHandler_SaveJSON(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	rtr := mux.NewRouter()
	c := configs.Instance()
	// Allocation handler and storage
	h := New(c.Logger, c.Storage)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "{\"url\": \"http://bench" + strconv.Itoa(i) + ".ru\"}"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer() //
		rtr.HandleFunc("/", h.Get)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}
