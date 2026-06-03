package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"AggregateNewsSF/pkg/api"
	"AggregateNewsSF/pkg/models"
	"AggregateNewsSF/pkg/rss"
	"AggregateNewsSF/pkg/storage"
)

func main() {
	config := readConfig("cmd/server/config.json")

	db, err := storage.New("postgres://postgres:Qwerty123@localhost:5432/aggregatenewssf")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	go startRSSParser(db, config)

	apiServer := api.New(db)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", apiServer.Router()))
}

func readConfig(filename string) *models.Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Failed to open config:", err)
	}
	defer file.Close()

	var config models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Failed to decode config:", err)
	}

	return &config
}

func startRSSParser(db *storage.Store, config *models.Config) {
	ticker := time.NewTicker(time.Duration(config.RequestPeriod) * time.Minute)

	for {
		var wg sync.WaitGroup
		postsChan := make(chan []models.Post, len(config.RSS))
		errorsChan := make(chan error, len(config.RSS))

		for _, rssURL := range config.RSS {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				log.Printf("Parsing RSS: %s", url)

				posts, err := rss.ParseRSS(url)
				if err != nil {
					errorsChan <- err
					return
				}
				postsChan <- posts
			}(rssURL)
		}

		wg.Wait()
		close(postsChan)
		close(errorsChan)

		for err := range errorsChan {
			log.Printf("RSS parsing error: %v", err)
		}

		for posts := range postsChan {
			if err := db.SavePosts(posts); err != nil {
				log.Printf("Failed to save posts: %v", err)
			} else {
				log.Printf("Saved %d posts", len(posts))
			}
		}

		<-ticker.C
	}
}
