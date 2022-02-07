package handlers

import (
	"fmt"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"io"
	"net/http"
	"regexp"
)

var regexpID = regexp.MustCompile(`^\/([\w\d]+)\/?$`)

type Shortener struct {
	Storage repositories.Shortener
}

func (h *Shortener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		link, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := h.Storage.Add(string(link))
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "http://localhost:8080/%v", id)
	case r.Method == http.MethodGet && regexpID.MatchString(r.URL.Path):
		id := regexpID.FindStringSubmatch(r.URL.Path)[1]

		link, err := h.Storage.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, link, http.StatusTemporaryRedirect)
	default:
		http.Error(w, "expect method GET /{id} or POST /", http.StatusBadRequest)
	}
}
