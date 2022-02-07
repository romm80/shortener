package handlers

import (
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortener_ServeHTTP(t *testing.T) {
	storage := &repositories.MapStorage{}
	storage.Init()
	handler := &Shortener{Storage: storage}

	type want struct {
		statusCode int
		location   string
		body       string
	}
	tests := []struct {
		name    string
		request string
		body    string
		method  string
		storage repositories.Shortener
		want    want
	}{
		// TODO: Add test cases.
		{
			name:    "Add link 1",
			request: "/",
			body:    "https://www.yandex.ru",
			method:  http.MethodPost,
			storage: storage,
			want: want{
				statusCode: 201,
				location:   "",
				body:       "http://localhost:8080/1"}},
		{
			name:    "Add link 2",
			request: "/",
			body:    "https://www.google.com",
			method:  http.MethodPost,
			storage: storage,
			want: want{
				statusCode: 201,
				location:   "",
				body:       "http://localhost:8080/2"}},
		{
			name:    "Get link 1",
			request: "/1",
			method:  http.MethodGet,
			storage: storage,
			want: want{
				statusCode: 307,
				location:   "https://www.yandex.ru"}},
		{
			name:    "Get link 2",
			request: "/2",
			method:  http.MethodGet,
			storage: storage,
			want: want{
				statusCode: 307,
				location:   "https://www.google.com"}},
		{
			name:    "Not found link ID",
			request: "/3",
			method:  http.MethodGet,
			storage: storage,
			want: want{
				statusCode: 400,
				body:       "not found link id\n"}},
		{
			name:    "Wrong method 1",
			request: "/",
			method:  http.MethodGet,
			storage: storage,
			want: want{
				statusCode: 400,
				body:       "expect method GET /{id} or POST /\n"}},
		{
			name:    "Wrong method 2",
			request: "/1/2",
			method:  http.MethodGet,
			storage: storage,
			want: want{
				statusCode: 400,
				body:       "Expect method GET /{id} or POST /\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			request := httptest.NewRequest(tt.method, tt.request, body)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))

			if tt.method == http.MethodPost {
				link, err := ioutil.ReadAll(result.Body)
				require.NoError(t, err)
				err = result.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, tt.want.body, string(link))
			}
		})
	}
}
