package main

import (
	"errors"
	"net/http"

	"github.com/manuelam2003/blogly/internal/data"
)

func (app *application) addPostTagHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	tagID, err := app.readIDParam(r, "tag_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.PostTags.Insert(postID, tagID)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateEntry) {
			app.conflictResponse(w, r, err)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "tag successfully added to post"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deletePostTagHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	tagID, err := app.readIDParam(r, "tag_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.PostTags.Delete(postID, tagID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "tag successfully removed from post"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
