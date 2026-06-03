package models

import "time"

type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	PubTime int64  `json:"pub_time"`
	Link    string `json:"link"`
}

func (p *Post) FormattedTime() string {
	return time.Unix(p.PubTime, 0).UTC().Format("02.01.2006 15:04:05")
}

type Config struct {
	RSS           []string `json:"rss"`
	RequestPeriod int      `json:"request_period"`
}

type RSSFeed struct {
	Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Items       []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Content     string `xml:"encoded"`
		} `xml:"item"`
	} `xml:"channel"`
}
