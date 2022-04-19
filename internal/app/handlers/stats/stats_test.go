package stats

import (
	"github.com/stretchr/testify/assert"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	l := zap.NewNop()
	c := configs.Instance()
	h := NewStats(c.Storage, l)

	w := httptest.NewRecorder()
	r := strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", r)

	h.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode, "не верный код ответа")

	defer res.Body.Close()
}

func TestNewStats(t *testing.T) {
	l := zap.NewNop()
	c := configs.Instance()
	h := NewStats(c.Storage, l)

	assert.IsType(t, &Handler{c.Storage, nil}, h)
}
