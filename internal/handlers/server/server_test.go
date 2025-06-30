package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//----------------------Test-Handlers----------------------

// NOTE: add to name that it is gaugeHandler in makeHandler()
func TestGaugeHandlerInWrapper(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	type request struct {
		method      string
		url         string
		contentType string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "method-get",
			request: request{
				method: http.MethodGet,
				url:    "/update/gauge/meow/12",
			},
			want: want{
				code: 405,
			},
		},
		{
			name: "no-name#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//3",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-val#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name//",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name//fff",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/fff",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "working-case#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/12.1",
			},
			want: want{
				code: 200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			MakeHandler(GaugeHandler)(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestCounterHandlerInWrapper(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	type request struct {
		method      string
		url         string
		contentType string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "method-get",
			request: request{
				method: http.MethodGet,
				url:    "/update/counter/meow/12",
			},
			want: want{
				code: 405,
			},
		},
		{
			name: "no-name#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter//",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-name#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter//3",
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "no-val#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name/",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name//",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name//fff",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "no-val#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name/fff",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "working-case#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/counter/1",
			},
			want: want{
				code: 200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			MakeHandler(CounterHandler)(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGeneralCaseHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	type request struct {
		method      string
		url         string
		contentType string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "no-type",
			request: request{
				method: http.MethodGet,
				url:    "/update",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "incorr-type",
			request: request{
				method: http.MethodGet,
				url:    "/update/name",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "incorr-path",
			request: request{
				method: http.MethodGet,
				url:    "/name/",
			},
			want: want{
				code: 400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			GeneralCaseHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
