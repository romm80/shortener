package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
	"github.com/romm80/shortener.git/internal/app/service/workers"
)

// @title           Shortener API
// @version         1.0
// @description     Link shortening service.

// @host      localhost:8080

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
	r.DeleteWorker.Run(r.Storage)

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
	r.Router.Use(r.TrustedIP)
	r.Router.GET("/api/internal/stats", r.DeleteUserURLs)

	pprof.Register(r.Router)
	return r, nil
}

// Add godoc
// @Summary      Adds a link
// @Description  Shortens the received link and adds it to the database
// @Accept       plain
// @Produce      plain
// @Param RequestURL body string true "original link"
// @Success 201 {string} string "short link"
// @Failure 500 {string} string "internal error"
// @Router       / [post]
func (s *Shortener) Add(c *gin.Context) {
	originURL, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	urlID, err := s.Storage.Add(string(originURL), c.GetUint64("userid"))
	statusCode := http.StatusCreated
	if err != nil && !errors.Is(err, app.ErrConflictURLID) {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if errors.Is(err, app.ErrConflictURLID) {
		statusCode = app.ErrStatusCode(err)
	}

	c.String(statusCode, "%s/%s", server.Cfg.BaseURL, urlID)
}

// AddJSON godoc
// @Summary      Adds a link
// @Description  Shortens the received link and adds it to the database
// @Accept       json
// @Produce      json
// @Param RequestURL body models.RequestURL true "original link"
// @Success 201 {object} models.ResponseURL "short link"
// @Failure 400 {string} string "invalid request"
// @Failure 500 {string} string "internal error"
// @Router       /api/shorten [post]
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
	if err != nil && !errors.Is(err, app.ErrConflictURLID) {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if errors.Is(err, app.ErrConflictURLID) {
		statusCode = app.ErrStatusCode(err)
	}

	res := models.ResponseURL{Result: fmt.Sprintf("%s/%s", server.Cfg.BaseURL, urlID)}
	c.JSON(statusCode, res)
}

// Get godoc
// @Summary      Redirects via a shortened link to the original
// @Description  Redirects via a shortened link to the original
// @Param id path string true "Link ID"
// @Success 307	{string} string "successfully redirected"
// @Failure 400 {string} string "Link not found"
// @Failure 410 {string} string "Link removed"
// @Failure 500 {string} string "internal error"
// @Router /{id} [get]
func (s *Shortener) Get(c *gin.Context) {
	urlID := c.Param("id")
	originURL, err := s.Storage.Get(urlID)
	if err != nil && !errors.Is(err, app.ErrDeletedURL) && !errors.Is(err, app.ErrLinkNoFound) {
		c.AbortWithStatus(app.ErrStatusCode(err))
		return
	}
	if errors.Is(err, app.ErrDeletedURL) {
		c.AbortWithStatus(app.ErrStatusCode(err))
		return
	}
	if errors.Is(err, app.ErrLinkNoFound) {
		c.AbortWithStatus(app.ErrStatusCode(err))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originURL)
}

// BatchURLs godoc
// @Summary      adds a links batch
// @Description  Shortens the link batch and adds it to the database
// @Accept       json
// @Produce      json
// @Param RequestURL body []models.RequestBatch true "original links"
// @Success 201 {object} []models.ResponseBatch "short links"
// @Failure 400 {string} string "invalid request"
// @Failure 409 {string} string "added link is already exist"
// @Failure 500 {string} string "internal error"
// @Router       /api/shorten/batch [post]
func (s *Shortener) BatchURLs(c *gin.Context) {
	reqBatch := make([]models.RequestBatch, 0)
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBatch); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	respBatch, err := s.Storage.AddBatch(reqBatch, c.GetUint64("userid"))
	if err != nil && !errors.Is(err, app.ErrConflictURLID) {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if errors.Is(err, app.ErrConflictURLID) {
		c.AbortWithError(http.StatusConflict, err)
		return
	}

	c.JSON(http.StatusCreated, respBatch)
}

// GetUserURLs godoc
// @Summary      Returns a list of links added by the user
// @Description  Returns a list of links added by the user
// @Accept       json
// @Produce      json
// @Success 200 {object} []models.UserURLs
// @Success 204 {string} string "user has no links"
// @Failure 500 {string} string "internal error"
// @Router       /api/user/urls [get]
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

// PingDB godoc
// @Summary      Checking the database connection
// @Description  Checking the database connection
// @Success 200 {string} string "successful connection"
// @Failure 500 {string} string "internal error"
// @Router       /ping [get]
func (s *Shortener) PingDB(c *gin.Context) {
	if err := s.Storage.Ping(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// DeleteUserURLs godoc
// @Summary      Removes user links by shortened ID
// @Description  Removes user links by shortened ID
// @Accept       json
// @Produce      json
// @Param urlsID body []string true "Link IDs to remove"
// @Success 202 {string} string "request accepted for processing"
// @Success 400 {string} string "invalid request"
// @Router       /api/user/urls [post]
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

func (s *Shortener) Stats(c *gin.Context) {
	res, err := s.Storage.GetStats()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}
