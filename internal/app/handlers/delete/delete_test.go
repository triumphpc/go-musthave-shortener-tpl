package delete

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ExampleHandler_ServeHTTP() {
	l := zap.NewNop()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	// Pool workers
	p, _ := worker.New(context.Background(), l, rep)

	h := New(l, p)
	w := httptest.NewRecorder()
	r := strings.NewReader("[\"123\"]")
	request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

	h.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()

}

func TestHandler_ServeHTTP(t *testing.T) {
	l := zap.NewNop()
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	// Pool workers
	p, poolClose := worker.New(context.Background(), l, rep)

	h := New(l, p)
	w := httptest.NewRecorder()
	r := strings.NewReader("[\"123\",\"555\", \"43433\", \"43234423\"]")
	request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

	h.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusAccepted, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()

	poolClose()
}

func ExampleNew() {
	l := zap.NewNop()
	_ = New(l, nil)
}

func TestNew(t *testing.T) {
	l := zap.NewNop()
	h := New(l, nil)

	assert.IsType(t, &Handler{l, nil}, h)

}
