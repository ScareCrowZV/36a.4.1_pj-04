package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AggregateNewsSF/pkg/models"
)

type MockStorage struct {
	posts []models.Post
	err   error
}

func (m *MockStorage) GetLastPosts(n int) ([]models.Post, error) {
	if m.err != nil {
		return nil, m.err
	}
	if n > len(m.posts) {
		return m.posts, nil
	}
	return m.posts[:n], nil
}

func (m *MockStorage) SavePosts(posts []models.Post) error {
	return nil
}

func (m *MockStorage) Close() {}

func setupTestTemplate(t *testing.T) {
	siteDir := filepath.Join(".", "site")
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		t.Logf("Не удалось создать директорию site: %v", err)
		return
	}

	tmplPath := filepath.Join(siteDir, "index.html")
	tmplContent := `<!DOCTYPE html>
<html>
<head>
    <title>AggregateNewsSF</title>
</head>
<body>
    <h1>AggregateNewsSF</h1>
    <div>{{.Now}}</div>
    {{range .Posts}}
        <div class="post">
            <a href="{{.Link}}">{{.Title}}</a>
            <div>{{.FormattedTime}}</div>
            <div>{{.Content}}</div>
        </div>
    {{end}}
</body>
</html>`

	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Logf("Не удалось записать шаблон: %v", err)
	}
}

func cleanupTestTemplate(t *testing.T) {
	tmplPath := filepath.Join("site", "index.html")
	os.Remove(tmplPath)
	os.Remove("site")
}

func TestNew(t *testing.T) {
	mockStorage := &MockStorage{}
	api := New(mockStorage)

	if api == nil {
		t.Error("Ожидался экземпляр API, получен nil")
	}
	if api.storage == nil {
		t.Error("Ожидалось хранилище, получено nil")
	}
}

func TestGetNews(t *testing.T) {
	mockStorage := &MockStorage{
		posts: []models.Post{
			{ID: 1, Title: "Post 1", Content: "Content 1", PubTime: 1640995200, Link: "https://example.com/1"},
			{ID: 2, Title: "Post 2", Content: "Content 2", PubTime: 1640995201, Link: "https://example.com/2"},
		},
	}

	api := New(mockStorage)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCount  int
	}{
		{"валидный запрос с 1", "/news/1", http.StatusOK, 1},
		{"валидный запрос с 2", "/news/2", http.StatusOK, 2},
		{"неверное число", "/news/invalid", http.StatusBadRequest, 0},
		{"отрицательное число", "/news/-5", http.StatusBadRequest, 0},
		{"ноль", "/news/0", http.StatusBadRequest, 0},
		{"большое число", "/news/200", http.StatusOK, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			api.getNews(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var posts []models.Post
				err := json.NewDecoder(w.Body).Decode(&posts)
				if err != nil {
					t.Errorf("Ошибка при декодировании ответа: %v", err)
				}
			}
		})
	}
}

func TestGetNewsMethodNotAllowed(t *testing.T) {
	mockStorage := &MockStorage{}
	api := New(mockStorage)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/news/5", nil)
			w := httptest.NewRecorder()

			api.getNews(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Для метода %s ожидался статус 405, получен %d", method, w.Code)
			}
		})
	}
}

func TestGetNewsStorageError(t *testing.T) {
	mockStorage := &MockStorage{
		err: &storageError{},
	}
	api := New(mockStorage)

	req := httptest.NewRequest("GET", "/news/5", nil)
	w := httptest.NewRecorder()

	api.getNews(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Ожидался статус 500, получен %d", w.Code)
	}
}

type storageError struct{}

func (e *storageError) Error() string {
	return "ошибка хранилища"
}

func TestIndex(t *testing.T) {
	setupTestTemplate(t)
	defer cleanupTestTemplate(t)

	mockStorage := &MockStorage{
		posts: []models.Post{
			{ID: 1, Title: "Test Post", Content: "Test Content", PubTime: time.Now().Unix(), Link: "https://example.com"},
		},
	}
	api := New(mockStorage)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	api.index(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Тело ответа: %s", w.Body.String())
		t.Errorf("Ожидался статус 200, получен %d", w.Code)
	}
}

func TestIndexNotFound(t *testing.T) {
	mockStorage := &MockStorage{}
	api := New(mockStorage)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	api.index(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус 404, получен %d", w.Code)
	}
}

func TestIndexMethodNotAllowed(t *testing.T) {
	mockStorage := &MockStorage{}
	api := New(mockStorage)

	methods := []string{"POST", "PUT", "DELETE"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			api.index(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Для метода %s ожидался статус 405, получен %d", method, w.Code)
			}
		})
	}
}

func TestRouter(t *testing.T) {
	setupTestTemplate(t)
	defer cleanupTestTemplate(t)

	mockStorage := &MockStorage{
		posts: []models.Post{
			{ID: 1, Title: "Test", Content: "Content", PubTime: time.Now().Unix(), Link: "https://example.com"},
		},
	}
	api := New(mockStorage)
	handler := api.Router()

	if handler == nil {
		t.Error("Ожидался не nil обработчик")
	}

	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
	}{
		{"главная страница", "/", "GET", http.StatusOK},
		{"API новостей", "/news/5", "GET", http.StatusOK},
		{"неверный путь", "/invalid", "GET", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.wantStatus, w.Code)
			}
		})
	}
}
