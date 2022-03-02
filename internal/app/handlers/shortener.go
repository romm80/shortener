package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
	"io/ioutil"
	"net/http"
)

type Shortener struct {
	Router  *gin.Engine
	Storage repositories.Shortener
}

type OriginURL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	Result string `json:"result"`
}

func New() (*Shortener, error) {
	r := &Shortener{}
	s, err := mapstorage.New()
	if err != nil {
		return nil, err
	}
	r.Storage = s

	r.Router = gin.Default()
	r.Router.POST("/", r.Add)
	r.Router.GET("/:id", r.Get)
	r.Router.POST("/api/shorten", r.AddJSON)

	return r, nil
}

func (s *Shortener) Add(c *gin.Context) {
	link, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	id, err := s.Storage.Add(string(link))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, "%s/%s", server.Cfg.BaseURL, id)

}

func (s *Shortener) AddJSON(c *gin.Context) {
	var url OriginURL
	if err := json.NewDecoder(c.Request.Body).Decode(&url); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if url.URL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := s.Storage.Add(url.URL)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res := ShortURL{Result: fmt.Sprintf("%s/%s", server.Cfg.BaseURL, id)}
	c.JSON(http.StatusCreated, res)
}

func (s *Shortener) Get(c *gin.Context) {
	id := c.Param("id")
	link, err := s.Storage.Get(id)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, link)
}
