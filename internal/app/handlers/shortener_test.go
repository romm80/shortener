package handlers

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/server"
	"github.com/romm80/shortener.git/internal/app/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var urls = []models.URLsID{
	{
		ID:          service.ShortenURLID("https://www.google.com/"),
		OriginalURL: "https://www.google.com/",
	},
	{
		ID:          service.ShortenURLID("https://yandex.ru/"),
		OriginalURL: "https://yandex.ru/",
	},
}

func TestShortener_Add(t *testing.T) {
	server.Cfg.DBType = server.DBMap
	server.Cfg.FileStorage = ""

	handler, err := New()
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&server.Cfg); err != nil {
		log.Fatal(err)
	}

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
			body: urls[0].OriginalURL,
			want: want{
				status: 201,
				body:   service.BaseURL(urls[0].ID),
			},
		},
		{
			name: "Successfully added link 2",
			path: "/",
			body: urls[1].OriginalURL,
			want: want{
				status: 201,
				body:   service.BaseURL(urls[1].ID),
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
	server.Cfg.DBType = server.DBMap
	server.Cfg.FileStorage = ""
	handler, err := New()
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&server.Cfg); err != nil {
		log.Fatal(err)
	}

	userID, _ := handler.Storage.NewUser()
	urls := []models.URLsID{
		{
			OriginalURL: "https://www.google.com/",
		},
		{
			OriginalURL: "https://yandex.ru/",
		},
	}

	urls[0].ID, _ = handler.Storage.Add(urls[0].OriginalURL, userID)
	urls[1].ID, _ = handler.Storage.Add(urls[1].OriginalURL, userID)

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
			id:   urls[0].ID,
			want: want{
				status:   307,
				location: urls[0].OriginalURL,
			},
		},
		{
			name: "Successfully received link 2",
			path: "/",
			id:   urls[1].ID,
			want: want{
				status:   307,
				location: urls[1].OriginalURL,
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
			result.Body.Close()

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func TestShortener_AddJSON(t *testing.T) {
	server.Cfg.DBType = server.DBMap

	handler, err := New()
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&server.Cfg); err != nil {
		log.Fatal(err)
	}

	type want struct {
		status      int
		contentType string
		body        string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Successfully added link 1",
			path: "/api/shorten",
			body: fmt.Sprintf(`{"url":"%s"}`, urls[0].OriginalURL),
			want: want{
				status:      201,
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"result":"%s"}`, service.BaseURL(urls[0].ID)),
			},
		},
		{
			name: "Successfully added link 2",
			path: "/api/shorten",
			body: fmt.Sprintf(`{"url":"%s"}`, urls[1].OriginalURL),
			want: want{
				status:      201,
				contentType: "application/json; charset=utf-8",
				body:        fmt.Sprintf(`{"result":"%s"}`, service.BaseURL(urls[1].ID)),
			},
		},
		{
			name: "invalid json 1",
			path: "/api/shorten",
			body: `{"url2":"https://yandex.ru/"}`,
			want: want{
				status: 400,
			},
		},
		{
			name: "invalid json 2",
			path: "/api/shorten",
			body: `{"url2":1}`,
			want: want{
				status: 400,
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
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			link, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(link))
		})
	}
}
