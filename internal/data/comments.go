package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/manuelam2003/blogly/internal/validator"
)

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentModel struct {
	DB *sql.DB
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.PostID > 0, "post_id", "must be greater than zero")
	v.Check(comment.UserID > 0, "user_id", "must be greater than zero")

	v.Check(comment.Content != "", "content", "must be provided")
	v.Check(len(comment.Content) <= 3000, "content", "must not be more than 3000 bytes long")
}

func (m CommentModel) Get(postID, commentID int64) (*Comment, error) {
	query := `
        SELECT id, post_id, user_id, content, created_at, updated_at
        FROM comments
        WHERE post_id = $1 AND id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var comment Comment

	err := m.DB.QueryRowContext(ctx, query, postID, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &comment, nil
}

func (c CommentModel) GetAllForPost(postID int64, filters Filters) ([]*Comment, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, post_id, user_id, content, created_at, updated_at
        FROM comments
        WHERE post_id = $1
        ORDER BY %s %s
        LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, postID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var comments []*Comment

	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return comments, metadata, nil
}

func (c CommentModel) GetAllByUser(userID int64, filters Filters) ([]*Comment, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, post_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE user_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	args := []any{userID, filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	comments := []*Comment{}

	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		comments = append(comments, &comment)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return comments, metadata, nil
}

func (c CommentModel) Insert(comment *Comment) error {
	query := `
	INSERT INTO comments(post_id, user_id, content)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`

	args := []any{comment.PostID, comment.UserID, comment.Content}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (c CommentModel) Update(comment *Comment) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2 AND updated_at = $3
		RETURNING updated_at`

	args := []any{comment.Content, comment.ID, comment.UpdatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.UpdatedAt)

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

func (c CommentModel) Delete(id, userID, postID int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM comments
        WHERE id = $1 AND user_id = $2 AND post_id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id, userID, postID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		existsQuery := `
            SELECT user_id 
            FROM comments 
            WHERE id = $1 AND post_id = $2`

		var ownerID int64
		err := c.DB.QueryRowContext(ctx, existsQuery, id, postID).Scan(&ownerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrRecordNotFound
			}
			return err
		}

		if ownerID != userID {
			return ErrUnauthorized
		}
	}

	return nil
}
