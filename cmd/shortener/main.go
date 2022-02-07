package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

var (
	links = make(map[string]string)
	IDre  = regexp.MustCompile(`^\/([\w\d]+)\/?$`)
	mu    = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		link, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu.Lock()
		id := strconv.Itoa(len(links) + 1)
		links[id] = string(link)
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "http://localhost:8080/%v", id)
	case r.Method == http.MethodGet && IDre.MatchString(r.URL.Path):
		id := IDre.FindStringSubmatch(r.URL.Path)[1]
		if link, ok := links[id]; ok {
			http.Redirect(w, r, link, http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "Not found link ID", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Expect method GET /{id} or POST", http.StatusBadRequest)
	}
}
