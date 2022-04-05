package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
	"github.com/romm80/shortener.git/internal/app/service/workers"
	"io/ioutil"
	"net/http"
)

type Shortener struct {
	Router       *gin.Engine
	Storage      repositories.Shortener
	DeleteWorker *workers.DeleteWorker
}

func New() (*Shortener, error) {
	r := &Shortener{DeleteWorker: workers.NewDeleteWorker(1000)}
	var err error
	if r.Storage, err = repositories.NewStorage(); err != nil {
		return nil, err
	}
	r.DeleteWorker.Run(r.Storage.DeleteBatch)

	r.Router = gin.Default()
	r.Router.GET("/ping", r.PingDB)
	r.Router.Use(GzipMiddleware)
	r.Router.GET("/:id", r.Get)
	r.Router.Use(r.AuthMiddleware)
	r.Router.POST("/", r.Add)
	r.Router.POST("/api/shorten", r.AddJSON)
	r.Router.POST("/api/shorten/batch", r.BatchURLs)
	r.Router.GET("/api/user/urls", r.GetUserURLs)
	r.Router.DELETE("/api/user/urls", r.DeleteUserURLs)

	return r, nil
}

func (s *Shortener) Add(c *gin.Context) {
	originURL, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	urlID, err := s.Storage.Add(string(originURL), c.GetUint64("userid"))
	statusCode := http.StatusCreated
	if err != nil {
		if errors.Is(err, app.ErrConflictURLID) {
			statusCode = http.StatusConflict
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	c.String(statusCode, "%s/%s", server.Cfg.BaseURL, urlID)
}

func (s *Shortener) AddJSON(c *gin.Context) {
	var request models.RequestURL
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if request.URL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	urlID, err := s.Storage.Add(request.URL, c.GetUint64("userid"))
	statusCode := http.StatusCreated
	if err != nil {
		if errors.Is(err, app.ErrConflictURLID) {
			statusCode = http.StatusConflict
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	res := models.ResponseURL{Result: fmt.Sprintf("%s/%s", server.Cfg.BaseURL, urlID)}
	c.JSON(statusCode, res)
}

func (s *Shortener) Get(c *gin.Context) {
	urlID := c.Param("id")
	originURL, err := s.Storage.Get(urlID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, app.ErrDeletedURL) {
			status = http.StatusGone
		}
		c.AbortWithStatus(status)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, originURL)
}

func (s *Shortener) BatchURLs(c *gin.Context) {
	reqBatch := make([]models.RequestBatch, 0)
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBatch); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	respBatch, err := s.Storage.AddBatch(reqBatch, c.GetUint64("userid"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, respBatch)
}
func (s *Shortener) GetUserURLs(c *gin.Context) {
	userID := c.GetUint64("userid")
	res, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if len(res) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Shortener) PingDB(c *gin.Context) {
	if err := s.Storage.Ping(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func (s *Shortener) DeleteUserURLs(c *gin.Context) {
	urlsID := make([]string, 0)
	if err := c.BindJSON(&urlsID); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if len(urlsID) == 0 {
		c.AbortWithError(http.StatusBadRequest, app.ErrEmptyRequest)
		return
	}

	userID := c.GetUint64("userid")
	s.DeleteWorker.Add(userID, urlsID)

	c.Status(http.StatusAccepted)
}
