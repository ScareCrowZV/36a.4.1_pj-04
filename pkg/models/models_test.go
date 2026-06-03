package models

import (
	"testing"
	"time"
)

func TestPost_FormattedTime(t *testing.T) {
	tests := []struct {
		name     string
		pubTime  int64
		expected string
	}{
		{
			name:     "валидная временная метка",
			pubTime:  1640995200,
			expected: "01.01.2022 00:00:00",
		},
		{
			name:     "нулевая временная метка",
			pubTime:  0,
			expected: "01.01.1970 00:00:00",
		},
		{
			name:     "текущее время",
			pubTime:  time.Now().UTC().Unix(),
			expected: time.Now().UTC().Format("02.01.2006 15:04:05"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := Post{PubTime: tt.pubTime}
			result := post.FormattedTime()
			if tt.name == "текущее время" {
				if result != tt.expected {
					t.Errorf("Ожидалось %s, получено %s", tt.expected, result)
				}
			} else if result != tt.expected {
				t.Errorf("Ожидалось %s, получено %s", tt.expected, result)
			}
		})
	}
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		RSS:           []string{"https://example.com/rss", "https://example2.com/rss"},
		RequestPeriod: 5,
	}

	if len(config.RSS) != 2 {
		t.Errorf("Ожидалось 2 RSS ленты, получено %d", len(config.RSS))
	}
	if config.RequestPeriod != 5 {
		t.Errorf("Ожидался период опроса 5, получено %d", config.RequestPeriod)
	}
}
