package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/service"
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
	app.Cfg.DBType = app.DBMap
	app.Cfg.FileStorage = ""

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := New(services)
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&app.Cfg); err != nil {
		log.Fatal(err)
	}

	type want struct {
		body   string
		status int
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
	app.Cfg.DBType = app.DBMap
	app.Cfg.FileStorage = ""

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := New(services)
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&app.Cfg); err != nil {
		log.Fatal(err)
	}

	userID, _ := handler.Services.NewUser()
	urls = []models.URLsID{
		{
			OriginalURL: "https://www.google.com/",
		},
		{
			OriginalURL: "https://yandex.ru/",
		},
	}

	urls[0].ID, _ = handler.Services.Add(urls[0].OriginalURL, userID)
	urls[1].ID, _ = handler.Services.Add(urls[1].OriginalURL, userID)

	type want struct {
		location string
		status   int
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
			id:   service.ShortenURLID(urls[0].OriginalURL),
			want: want{
				status:   307,
				location: urls[0].OriginalURL,
			},
		},
		{
			name: "Successfully received link 2",
			path: "/",
			id:   service.ShortenURLID(urls[1].OriginalURL),
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
	app.Cfg.DBType = app.DBMap

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := New(services)
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&app.Cfg); err != nil {
		log.Fatal(err)
	}

	type want struct {
		contentType string
		body        string
		status      int
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

func TestShortener_BatchURLs(t *testing.T) {
	app.Cfg.DBType = app.DBMap
	app.Cfg.BaseURL = "http://127.0.0.1:8080"
	batchPath := "/api/shorten/batch"
	RequestURLs := []models.RequestBatch{
		{
			CorrelationID: "-",
			OriginalURL:   "https://www.google.com/",
		},
		{
			CorrelationID: "-",
			OriginalURL:   "https://www.yandex.ru/",
		},
	}
	ResponseURLs := []models.ResponseBatch{
		{
			CorrelationID: "-",
			ShortURL:      service.BaseURL(service.ShortenURLID(RequestURLs[0].OriginalURL)),
		},
		{
			CorrelationID: "-",
			ShortURL:      service.BaseURL(service.ShortenURLID(RequestURLs[1].OriginalURL)),
		},
	}
	reqJSON, err := json.Marshal(RequestURLs)
	if err != nil {
		log.Fatal(err)
	}
	respJSON, err := json.Marshal(ResponseURLs)
	if err != nil {
		log.Fatal(err)
	}

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := New(services)
	if err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(&app.Cfg); err != nil {
		log.Fatal(err)
	}

	type want struct {
		contentType string
		body        string
		status      int
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Successfully added links",
			path: batchPath,
			body: string(reqJSON),
			want: want{
				status:      201,
				contentType: "application/json; charset=utf-8",
				body:        string(respJSON),
			},
		},
		{
			name: "invalid json",
			path: batchPath,
			body: `{"url2":"https://yandex.ru/"}`,
			want: want{
				status: 400,
				body:   "[]",
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

func TestShortener_GetUserURLs(t *testing.T) {
	app.Cfg.DBType = app.DBMap

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := New(services)
	if err != nil {
		log.Fatal(err)
	}
	if err = env.Parse(&app.Cfg); err != nil {
		log.Fatal(err)
	}
	userURLsPath := "/api/user/urls"
	URLs := []models.RequestBatch{
		{
			CorrelationID: "-",
			OriginalURL:   "https://www.google.com/",
		},
		{
			CorrelationID: "-",
			OriginalURL:   "https://www.yandex.ru/",
		},
	}
	userID, err := handler.Services.NewUser()
	if err != nil {
		log.Fatal(err)
	}
	shortURLs, err := handler.Services.AddBatch(URLs, userID)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := json.Marshal([]models.UserURLs{
		{
			OriginalURL: URLs[0].OriginalURL,
			ShortURL:    shortURLs[0].ShortURL,
		},
		{
			OriginalURL: URLs[1].OriginalURL,
			ShortURL:    shortURLs[1].ShortURL,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	type want struct {
		contentType string
		body        string
		status      int
	}
	tests := []struct {
		name   string
		path   string
		body   string
		want   want
		userID uint64
	}{
		{
			name:   "Successfully found links",
			path:   userURLsPath,
			userID: userID,
			want: want{
				status:      200,
				contentType: "application/json; charset=utf-8",
				body:        string(resp),
			},
		},
		{
			name:   "Successfully no content 1",
			path:   userURLsPath,
			userID: 999,
			want: want{
				status: 204,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodGet, tt.path, body)
			signedID, _ := service.SignUserID(tt.userID)
			cookie := &http.Cookie{
				Name:  "userid",
				Value: signedID,
			}
			request.AddCookie(cookie)
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
