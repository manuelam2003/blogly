package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/manuelam2003/blogly/internal/validator"
)

type Post struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ValidatePost(v *validator.Validator, post *Post) {
	v.Check(post.Title != "", "title", "must be provided")
	v.Check(len(post.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(post.Content != "", "content", "must be provided")
	v.Check(len(post.Content) <= 3000, "content", "must not be more than 3000 bytes long")
}

type PostModel struct {
	DB *sql.DB
}

func (p PostModel) Insert(post *Post) error {
	query := `
		INSERT INTO posts(user_id, title, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	args := []any{post.UserID, post.Title, post.Content}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (p PostModel) Get(id int64) (*Post, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM posts
		WHERE id = $1`

	var post Post

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (p PostModel) Update(post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3 AND updated_at = $4
		RETURNING updated_at
	`

	args := []any{post.Title, post.Content, post.ID, post.UpdatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, args...).Scan(&post.UpdatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (p PostModel) Delete(id int64, userID int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM posts
        WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		existsQuery := `
            SELECT COUNT(*) 
            FROM posts 
            WHERE id = $1`

		var count int
		err := p.DB.QueryRowContext(ctx, existsQuery, id).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			return ErrRecordNotFound
		} else {
			// Post exists but the user is not authorized
			return ErrUnauthorized
		}
	}

	return nil
}

func (p PostModel) GetAll(userID int64, title, content string, filters Filters) ([]*Post, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, user_id, title, content, created_at, updated_at
	FROM posts
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple',content) @@ plainto_tsquery('simple',$2) OR $2 = '')
	AND (user_id = $3 OR $3 = 0)
	ORDER BY %s %s, id ASC
	LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, content, userID, filters.limit(), filters.offset()}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	posts := []*Post{}

	for rows.Next() {
		var post Post

		err := rows.Scan(
			&totalRecords,
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return posts, metadata, nil
}

func (p PostModel) GetAllForUser(userID int64, filters Filters) ([]*Post, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, user_id, title, content, created_at, updated_at
	FROM posts
	WHERE (user_id = $1 OR $1 = 0)
	ORDER BY %s %s, id ASC
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{userID, filters.limit(), filters.offset()}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	posts := []*Post{}

	for rows.Next() {
		var post Post

		err := rows.Scan(
			&totalRecords,
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return posts, metadata, nil
}
