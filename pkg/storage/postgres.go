package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"

	"AggregateNewsSF/pkg/models"
)

type Storager interface {
	GetLastPosts(n int) ([]models.Post, error)
	GetPostsWithPagination(page, pageSize int, search string) ([]models.Post, int, error)
	GetPostsByID(id int) ([]models.Post, error)
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

func (s *Store) GetPostsWithPagination(page, pageSize int, search string) ([]models.Post, int, error) {
	offset := (page - 1) * pageSize

	var countQuery string
	var dataQuery string
	var countArgs []interface{}
	var dataArgs []interface{}

	if search != "" {
		countQuery = `
            SELECT COUNT(*) FROM posts 
            WHERE title ILIKE $1 OR content ILIKE $1
        `
		countArgs = append(countArgs, "%"+search+"%")

		dataQuery = `
            SELECT id, title, content, pub_time, link 
            FROM posts 
            WHERE title ILIKE $1 OR content ILIKE $1
            ORDER BY pub_time DESC 
            LIMIT $2 OFFSET $3
        `
		dataArgs = append(dataArgs, "%"+search+"%", pageSize, offset)
	} else {
		countQuery = `SELECT COUNT(*) FROM posts`

		dataQuery = `
            SELECT id, title, content, pub_time, link 
            FROM posts 
            ORDER BY pub_time DESC 
            LIMIT $1 OFFSET $2
        `
		dataArgs = append(dataArgs, pageSize, offset)
	}

	var total int
	err := s.db.QueryRow(context.Background(), countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	rows, err := s.db.Query(context.Background(), dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, p)
	}

	return posts, total, rows.Err()
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

func (s *Store) GetPostsByID(id int) ([]models.Post, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT id, title, content, pub_time, link FROM posts WHERE id = $1
	`, id)
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
