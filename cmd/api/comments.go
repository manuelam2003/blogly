package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/manuelam2003/blogly/internal/data"
	"github.com/manuelam2003/blogly/internal/validator"
)

func (app *application) listPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "created_at", "user_id", "-id", "-created_at", "-user_id"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	comments, metadata, err := app.models.Comments.GetAllForPost(postID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCommentHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	commentID, err := app.readIDParam(r, "comment_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	comment, err := app.models.Comments.Get(postID, commentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Content string `json:"content"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	currentUser := app.contextGetUser(r)

	comment := &data.Comment{
		PostID:  postID,
		UserID:  currentUser.ID,
		Content: input.Content,
	}

	v := validator.New()

	if data.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Comments.Insert(comment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/posts/%d/comments/%d", postID, comment.ID))

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	commentID, err := app.readIDParam(r, "comment_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	comment, err := app.models.Comments.Get(postID, commentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	currentUser := app.contextGetUser(r)

	if currentUser.ID != comment.UserID {
		app.invalidUserResponse(w, r)
		return
	}

	var input struct {
		Content *string `json:"content"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Content != nil {
		comment.Content = *input.Content
	}

	v := validator.New()

	if data.ValidateComment(v, comment); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Comments.Update(comment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	commentID, err := app.readIDParam(r, "comment_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	currentUser := app.contextGetUser(r)

	err = app.models.Comments.Delete(commentID, currentUser.ID, postID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, data.ErrUnauthorized):
			app.invalidUserResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "comment successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listUserCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := app.readIDParam(r, "user_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "created_at")
	input.Filters.SortSafelist = []string{"created_at", "updated_at", "-created_at", "-updated_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	comments, metadata, err := app.models.Comments.GetAllByUser(userID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
