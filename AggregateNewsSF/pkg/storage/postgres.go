package storage

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"AggregateNewsSF/pkg/models"
)

type Storager interface {
	GetLastPosts(n int) ([]models.Post, error)
	SavePosts(posts []models.Post) error
	Close()
}

type Store struct {
	db *pgxpool.Pool
}

func New(constr string) (*Store, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Store{
		db: db,
	}
	return &s, nil
}

func (s *Store) GetLastPosts(n int) ([]models.Post, error) {
	rows, err := s.db.Query(context.Background(), `
        SELECT id, title, content, pub_time, link FROM posts ORDER BY pub_time DESC LIMIT $1
    `, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *Store) SavePosts(posts []models.Post) error {
	for _, post := range posts {
		_, err := s.db.Exec(context.Background(), `
            INSERT INTO posts (title, content, pub_time, link)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (link) DO NOTHING
        `,
			post.Title,
			post.Content,
			post.PubTime,
			post.Link,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
