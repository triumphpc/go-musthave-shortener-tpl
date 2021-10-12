package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	rep, err := file.New("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rep)
	h := New(zap.NewNop(), rep)

	type want struct {
		code        int
		response    string
		contentType string
	}

	type request struct {
		method     string
		target     string
		body       string
		path       string
		saveParam  bool
		checkParam bool
	}

	// Structure of tests
	tests := []struct {
		name    string
		want    want
		request request
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		// implement all tests
		{
			name:    "Test increment #9",
			handler: h.GetUrls,
			request: request{
				method: http.MethodGet,
				target: "/user/urls",
				path:   "/user/urls",
			},
			want: want{
				code:        http.StatusNoContent,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test Save handler #1",
			handler: h.Save,
			request: request{
				method:    http.MethodPost,
				target:    "/",
				path:      "/",
				body:      "http://newlink.ru",
				saveParam: true,
			},
			want: want{
				code:        http.StatusCreated,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test Save handler #2",
			handler: h.Save,
			request: request{
				method: http.MethodPost,
				target: "/",
				path:   "/",
				body:   "",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test Get handler #1",
			handler: h.Get,
			request: request{
				method: http.MethodGet,
				target: "/xxx",
				path:   "/{id:.+}",
				body:   "",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test Get handler #2",
			handler: h.Get,
			request: request{
				method: http.MethodGet,
				target: "/GMWJGSAPGA_test_1",
				path:   "/{id:.+}",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test Get handler #3",
			handler: h.Get,
			request: request{
				method:     http.MethodGet,
				target:     "/",
				path:       "/{id:.+}",
				checkParam: true,
			},
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "",
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name:    "Test SaveJSON handler #1",
			handler: h.SaveJSON,
			request: request{
				method:    http.MethodPost,
				target:    "/",
				path:      "/",
				body:      "{\"urfxxxx\": \"http://vtest.com\"}",
				saveParam: true,
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "unknown url",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "Test SaveJSON handler #2",
			handler: h.SaveJSON,
			request: request{
				method:     http.MethodPost,
				target:     "/",
				path:       "/",
				body:       "{\"url\": \"http://test.ru\"}",
				saveParam:  false,
				checkParam: true,
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	type lastParams struct {
		link      string
		shortLink string
	}
	var lp lastParams

	for _, tt := range tests {
		// Run tests
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if len(tt.request.body) > 0 {
				r = strings.NewReader(tt.request.body)
			} else {
				r = nil
			}

			// Add to target request
			if tt.request.method == http.MethodGet {
				if tt.request.checkParam {
					tt.request.target = lp.shortLink
				}
			}

			request := httptest.NewRequest(tt.request.method, tt.request.target, r)

			// Create new recorder
			w := httptest.NewRecorder()
			// Init handler
			rtr := mux.NewRouter()

			rtr.HandleFunc(tt.request.path, tt.handler)

			// запускаем сервер
			rtr.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, tt.want.code, res.StatusCode, "не верный код ответа")

			// тело запроса
			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			readLine := strings.TrimSuffix(string(resBody), "\n")
			// equal response
			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, readLine)

			}

			if res.StatusCode == tt.want.code {
				assert.Positive(t, readLine)
				// Save last param
				if tt.request.saveParam {
					lp.link = tt.request.body
					lp.shortLink = readLine
				}
			}

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
