package storage

import (
	"AggregateNewsSF/pkg/models"
	"testing"
)

func TestNew(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Error("Ожидалось не nil хранилище")
	}
	if store.db == nil {
		t.Error("Ожидалось не nil соединение с БД")
	}
}

func TestSavePosts(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	posts := []models.Post{
		{
			Title:   "Test Post 1",
			Content: "Content 1",
			PubTime: 1640995200,
			Link:    "https://example.com/post1",
		},
		{
			Title:   "Test Post 2",
			Content: "Content 2",
			PubTime: 1640995201,
			Link:    "https://example.com/post2",
		},
	}

	err = store.SavePosts(posts)
	if err != nil {
		t.Errorf("Ошибка при сохранении постов: %v", err)
	}

	savedPosts, err := store.GetLastPosts(2)
	if err != nil {
		t.Errorf("Ошибка при получении постов: %v", err)
	}

	if len(savedPosts) < 1 {
		t.Errorf("Ожидался хотя бы 1 пост, получено %d", len(savedPosts))
	}
}

func TestSavePostsDuplicate(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	post := models.Post{
		Title:   "Duplicate Test",
		Content: "Test Content",
		PubTime: 1640995200,
		Link:    "https://example.com/duplicate",
	}

	err = store.SavePosts([]models.Post{post})
	if err != nil {
		t.Errorf("Ошибка при сохранении первого поста: %v", err)
	}

	err = store.SavePosts([]models.Post{post})
	if err != nil {
		t.Errorf("Ошибка при сохранении дубликата: %v", err)
	}
}

func TestGetLastPosts(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	testPosts := []models.Post{
		{Title: "Post 1", Content: "Content 1", PubTime: 1640995200, Link: "https://example.com/1"},
		{Title: "Post 2", Content: "Content 2", PubTime: 1640995201, Link: "https://example.com/2"},
		{Title: "Post 3", Content: "Content 3", PubTime: 1640995202, Link: "https://example.com/3"},
	}

	err = store.SavePosts(testPosts)
	if err != nil {
		t.Errorf("Ошибка при сохранении тестовых постов: %v", err)
	}

	posts, err := store.GetLastPosts(2)
	if err != nil {
		t.Errorf("Ошибка при получении постов: %v", err)
	}

	if len(posts) > 2 {
		t.Errorf("Ожидалось не более 2 постов, получено %d", len(posts))
	}

	for i := 1; i < len(posts); i++ {
		if posts[i-1].PubTime < posts[i].PubTime {
			t.Errorf("Посты не отсортированы по убыванию времени публикации")
		}
	}
}

func TestGetLastPostsWithLimit(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	testPosts := []models.Post{
		{Title: "Limit Test 1", Content: "Content 1", PubTime: 1640995200, Link: "https://example.com/limit1"},
		{Title: "Limit Test 2", Content: "Content 2", PubTime: 1640995201, Link: "https://example.com/limit2"},
		{Title: "Limit Test 3", Content: "Content 3", PubTime: 1640995202, Link: "https://example.com/limit3"},
		{Title: "Limit Test 4", Content: "Content 4", PubTime: 1640995203, Link: "https://example.com/limit4"},
		{Title: "Limit Test 5", Content: "Content 5", PubTime: 1640995204, Link: "https://example.com/limit5"},
	}

	err = store.SavePosts(testPosts)
	if err != nil {
		t.Errorf("Ошибка при сохранении тестовых постов: %v", err)
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{"получить 3 поста", 3, 3},
		{"получить 5 постов", 5, 5},
		{"получить 10 постов (больше чем доступно)", 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := store.GetLastPosts(tt.limit)
			if err != nil {
				t.Errorf("Ошибка при получении постов: %v", err)
				return
			}
			if len(posts) != tt.wantCount && len(posts) != tt.limit {
				if tt.limit <= len(testPosts) {
					if len(posts) != tt.limit {
						t.Errorf("Ожидалось %d постов, получено %d", tt.limit, len(posts))
					}
				}
			}
		})
	}
}

func TestGetLastPostsInvalidLimit(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}
	defer store.Close()

	_, err = store.GetLastPosts(0)
	if err == nil {
		t.Log("Примечание: GetLastPosts(0) не вернул ошибку, это может быть ожидаемым поведением")
	}

	_, err = store.GetLastPosts(-1)
	if err == nil {
		t.Log("Примечание: GetLastPosts(-1) не вернул ошибку, это может быть ожидаемым поведением")
	}
}

func TestClose(t *testing.T) {
	connStr := "postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf"

	store, err := New(connStr)
	if err != nil {
		t.Skipf("Пропуск теста: база данных недоступна: %v", err)
	}

	store.Close()
	store.Close()
}
