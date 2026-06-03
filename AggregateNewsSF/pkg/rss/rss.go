package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"AggregateNewsSF/pkg/models"
)

func ParseRSS(url string) ([]models.Post, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var feed models.RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var posts []models.Post
	for _, item := range feed.Channel.Items {
		pubTime := parsePubDate(item.PubDate)

		content := item.Description
		if content == "" {
			content = item.Content
		}

		post := models.Post{
			Title:   item.Title,
			Content: content,
			PubTime: pubTime,
			Link:    item.Link,
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func parsePubDate(dateStr string) int64 {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Unix()
		}
	}

	return time.Now().Unix()
}
