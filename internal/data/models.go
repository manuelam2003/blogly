package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrUnauthorized   = errors.New("user is not authorized to perform this action")
	ErrDuplicateEntry = errors.New("duplicate entry")
)

type Models struct {
	Posts    PostModel
	Users    UserModel
	Comments CommentModel
	Tags     TagModel
	PostTags PostTagModel
	Tokens   TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts:    PostModel{DB: db},
		Users:    UserModel{DB: db},
		Comments: CommentModel{DB: db},
		Tags:     TagModel{DB: db},
		PostTags: PostTagModel{DB: db},
		Tokens:   TokenModel{DB: db},
	}
}
