package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortener_Add(t *testing.T) {
	handler := New()

	type want struct {
		status int
		body   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Successfully added link 1",
			path: "/",
			body: "https://www.google.com/",
			want: want{
				status: 201,
				body:   "http://127.0.0.1:8080/d752",
			},
		},
		{
			name: "Successfully added link 2",
			path: "/",
			body: "https://yandex.ru/",
			want: want{
				status: 201,
				body:   "http://127.0.0.1:8080/30b7",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, tt.path, body)
			w := httptest.NewRecorder()

			handler.Router.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.status, result.StatusCode)

			link, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(link))
		})
	}
}

func TestShortener_Get(t *testing.T) {
	handler := New()

	type want struct {
		status   int
		location string
	}
	tests := []struct {
		name string
		path string
		id   string
		want want
	}{
		{
			name: "Successfully received link 1",
			path: "/",
			id:   handler.Storage.Add("https://www.google.com/"),
			want: want{
				status:   307,
				location: "https://www.google.com/",
			},
		},
		{
			name: "Successfully received link 2",
			path: "/",
			id:   handler.Storage.Add("https://yandex.ru/"),
			want: want{
				status:   307,
				location: "https://yandex.ru/",
			},
		},
		{
			name: "Link not found by id",
			path: "/",
			id:   "1234",
			want: want{
				status: 400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path+tt.id, nil)
			w := httptest.NewRecorder()

			handler.Router.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}