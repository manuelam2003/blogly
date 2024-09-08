package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/manuelam2003/blogly/internal/validator"
)

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ValidateTag(v *validator.Validator, tag *Tag) {
	v.Check(tag.Name != "", "name", "must be provided")
	v.Check(len(tag.Name) <= 500, "name", "must not be more than 500 bytes long")
}

type TagModel struct {
	DB *sql.DB
}

func (t TagModel) Insert(tag *Tag) error {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, tag.Name).Scan(&tag.ID, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "tags_name_key"`:
			return ErrDuplicateEntry
		default:
			return err
		}
	}

	return nil
}

func (t TagModel) Get(id int64) (*Tag, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, created_at, updated_at
		FROM tags
		WHERE id = $1`

	var tag Tag

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &tag, nil
}

func (t TagModel) Update(tag *Tag) error {
	query := `
		UPDATE tags
		SET name = $1, updated_at = NOW()
		WHERE id = $2 AND updated_at = $3
		RETURNING updated_at`

	args := []any{tag.Name, tag.ID, tag.UpdatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, args...).Scan(&tag.UpdatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "tags_name_key"`:
			return ErrDuplicateEntry
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (t TagModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM tags
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, id)
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

func (t TagModel) GetAllOld(name string, filters Filters) ([]*Tag, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id,name , created_at, updated_at
	FROM tags
	WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	ORDER BY %s %s, id ASC
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, filters.limit(), filters.offset()}

	rows, err := t.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	tags := []*Tag{}

	for rows.Next() {
		var tag Tag

		err := rows.Scan(
			&totalRecords,
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return tags, metadata, nil
}

func (t TagModel) GetAllForPost(postID int64, filters Filters) ([]*Tag, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), tags.id, tags.name, tags.created_at, tags.updated_at
	FROM tags
	INNER JOIN post_tags ON post_tags.tag_id = tags.id
	WHERE (post_tags.post_id = $1)
	ORDER BY %s %s, id ASC
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{postID, filters.limit(), filters.offset()}

	rows, err := t.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	tags := []*Tag{}

	for rows.Next() {
		var tag Tag

		err := rows.Scan(
			&totalRecords,
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return tags, metadata, nil
}

func (t TagModel) GetAll(postID int64, name string, filters Filters) ([]*Tag, Metadata, error) {
	var whereClause string
	var args []any

	if postID > 0 {
		whereClause = `INNER JOIN post_tags ON post_tags.tag_id = tags.id WHERE post_tags.post_id = $1`
		args = append(args, postID)
	} else {
		whereClause = `WHERE (to_tsvector('simple', tags.name) @@ plainto_tsquery('simple', $1) OR $1 = '')`
		args = append(args, name)
	}

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), tags.id, tags.name, tags.created_at, tags.updated_at
	FROM tags
	%s
	ORDER BY %s %s, tags.id ASC
	LIMIT $2 OFFSET $3`, whereClause, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Add pagination arguments
	args = append(args, filters.limit(), filters.offset())

	rows, err := t.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	tags := []*Tag{}

	for rows.Next() {
		var tag Tag

		err := rows.Scan(
			&totalRecords,
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return tags, metadata, nil
}
