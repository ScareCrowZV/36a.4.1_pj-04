package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseRSSWithMockServer(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS Feed</title>
    <link>https://example.com</link>
    <description>Test RSS Description</description>
    <item>
      <title>First Test Article</title>
      <link>https://example.com/article1</link>
      <description><![CDATA[<p>This is a <strong>test</strong> article with <a href="https://example.com">HTML</a></p>]]></description>
      <pubDate>Tue, 03 Jun 2025 12:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Second Test Article</title>
      <link>https://example.com/article2</link>
      <description><![CDATA[<p>Another test article with <em>formatting</em></p>]]></description>
      <pubDate>Tue, 03 Jun 2025 13:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Third Test Article with Encoded Content</title>
      <link>https://example.com/article3</link>
      <description>Simple description</description>
      <encoded><![CDATA[<p>This is encoded content with <strong>HTML tags</strong></p>]]></encoded>
      <pubDate>Tue, 03 Jun 2025 14:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockXML))
	}))
	defer server.Close()

	posts, err := ParseRSS(server.URL)
	if err != nil {
		t.Fatalf("Ошибка при парсинге тестовой RSS: %v", err)
	}

	if len(posts) != 3 {
		t.Errorf("Ожидалось 3 поста, получено %d", len(posts))
	}

	tests := []struct {
		index                   int
		expectedTitle           string
		expectedLink            string
		expectedContentContains string
	}{
		{0, "First Test Article", "https://example.com/article1", "<strong>test</strong>"},
		{1, "Second Test Article", "https://example.com/article2", "<em>formatting</em>"},
		{2, "Third Test Article with Encoded Content", "https://example.com/article3", "<strong>HTML tags</strong>"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedTitle, func(t *testing.T) {
			post := posts[tt.index]
			if post.Title != tt.expectedTitle {
				t.Errorf("Ожидался заголовок '%s', получено '%s'", tt.expectedTitle, post.Title)
			}
			if post.Link != tt.expectedLink {
				t.Errorf("Ожидалась ссылка '%s', получена '%s'", tt.expectedLink, post.Link)
			}
		})
	}
}

func TestParseRSSWithInvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is not valid XML"))
	}))
	defer server.Close()

	posts, err := ParseRSS(server.URL)
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного XML, получено nil")
	}
	if posts != nil {
		t.Error("Ожидался nil для невалидного XML")
	}
}

func TestParseRSSWithEmptyFeed(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>Empty RSS Feed</description>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockXML))
	}))
	defer server.Close()

	posts, err := ParseRSS(server.URL)
	if err != nil {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
	if len(posts) != 0 {
		t.Errorf("Ожидалось 0 постов, получено %d", len(posts))
	}
}
