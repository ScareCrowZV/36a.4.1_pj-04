package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"AggregateNewsSF/pkg/storage"
)

type API struct {
	storage storage.Storager
}

func New(storage storage.Storager) *API {
	api := &API{
		storage: storage,
	}
	return api
}

func (api *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/news/", api.getNews)
	mux.HandleFunc("/", api.index)
	return mux
}

func (api *API) getNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/news/")
	n, err := strconv.Atoi(path)
	if err != nil || n < 1 {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	if n > 100 {
		n = 100
	}

	posts, err := api.storage.GetLastPosts(n)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(posts)
}

func (api *API) index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts, err := api.storage.GetLastPosts(10)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("site", "index.html")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type PostWithHTML struct {
		ID            int
		Title         string
		Content       template.HTML
		PubTime       int64
		Link          string
		FormattedTime string
	}

	var postsWithHTML []PostWithHTML
	for _, p := range posts {
		postsWithHTML = append(postsWithHTML, PostWithHTML{
			ID:            p.ID,
			Title:         p.Title,
			Content:       template.HTML(p.Content),
			PubTime:       p.PubTime,
			Link:          p.Link,
			FormattedTime: p.FormattedTime(),
		})
	}

	data := struct {
		Posts []PostWithHTML
		Now   string
	}{
		Posts: postsWithHTML,
		Now:   time.Now().Format("02.01.2006 15:04:05"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}
