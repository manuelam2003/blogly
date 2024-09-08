package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/manuelam2003/blogly/internal/validator"
)

type PostTag struct {
	ID     int64 `json:"id"`
	PostID int64 `json:"post_id"`
	TagID  int64 `json:"tag_id"`
}

func ValidatePostTag(v *validator.Validator, postTag *PostTag) {
	v.Check(postTag.ID > 0, "id", "must be non negative")
	v.Check(postTag.PostID > 0, "post_id", "must be non negative")
	v.Check(postTag.TagID > 0, "tag_id", "must be non negative")
}

type PostTagModel struct {
	DB *sql.DB
}

func (p PostTagModel) Insert(postID, tagID int64) error {
	checkQuery := `
        SELECT COUNT(*) 
        FROM post_tags 
        WHERE post_id = $1 AND tag_id = $2`

	var count int
	err := p.DB.QueryRow(checkQuery, postID, tagID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrDuplicateEntry
	}

	insertQuery := `
        INSERT INTO post_tags (post_id, tag_id)
        VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = p.DB.ExecContext(ctx, insertQuery, postID, tagID)
	if err != nil {
		return err
	}

	return nil
}

func (pt *PostTagModel) Delete(postID, tagID int64) error {
	query := `
		DELETE FROM post_tags
		WHERE post_id = $1 AND tag_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := pt.DB.ExecContext(ctx, query, postID, tagID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
