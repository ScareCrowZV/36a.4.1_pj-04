package api

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"AggregateNewsSF/pkg/models"
	"AggregateNewsSF/pkg/storage"
)

type API struct {
	storage      storage.Storager
	templatePath string
}

type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	ItemsPerPage int `json:"items_per_page"`
	TotalItems   int `json:"total_items"`
}

type NewsListResponse struct {
	Posts      []models.Post `json:"posts"`
	Pagination Pagination    `json:"pagination"`
}

type contextKey string

const RequestIDKey contextKey = "request_id"

func New(storage storage.Storager) *API {
	api := &API{
		storage:      storage,
		templatePath: filepath.Join("site", "index.html"),
	}
	return api
}

func NewWithTemplate(storage storage.Storager, templatePath string) *API {
	api := &API{
		storage:      storage,
		templatePath: templatePath,
	}
	return api
}

func (api *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/news/", api.getNews)
	mux.HandleFunc("/api/news", api.getNewsAPI)
	mux.HandleFunc("/", api.index)
	return api.middlewareChain(mux)
}

func (api *API) middlewareChain(next http.Handler) http.Handler {
	return api.requestIDMiddleware(api.loggingMiddleware(next))
}

func (api *API) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("request_id")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		requestID := r.Context().Value(RequestIDKey)
		if requestID == nil {
			requestID = "unknown"
		}

		log.Printf("[%s] %s %s %d %s | IP: %s | RequestID: %v",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
			r.RemoteAddr,
			requestID,
		)
	})
}

func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (api *API) getNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/news/")

	if path == "" || path == "api" {
		api.getNewsAPI(w, r)
		return
	}

	// Пытаемся получить ID новости
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	// Получаем одну новость по ID
	posts, err := api.storage.GetPostsByID(id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		http.Error(w, "News not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(posts[0])
}

func (api *API) getNewsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	search := r.URL.Query().Get("s")
	pageSize := 15

	posts, total, err := api.storage.GetPostsWithPagination(page, pageSize, search)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	response := NewsListResponse{
		Posts: posts,
		Pagination: Pagination{
			CurrentPage:  page,
			TotalPages:   totalPages,
			ItemsPerPage: pageSize,
			TotalItems:   total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
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

	tmpl, err := template.ParseFiles(api.templatePath)
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
