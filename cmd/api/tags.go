package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/manuelam2003/blogly/internal/data"
	"github.com/manuelam2003/blogly/internal/validator"
)

func (app *application) listTagsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "updated_at", "-id", "-name", "-updated_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	tags, metadata, err := app.models.Tags.GetAll(0, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tags": tags, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showTagHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r, "tag_id")
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	tag, err := app.models.Tags.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tag": tag}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createTagHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// TODO: hacer algo con esto
	// currentUser := app.contextGetUser(r)

	tag := &data.Tag{
		Name: input.Name,
	}

	v := validator.New()

	if data.ValidateTag(v, tag); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Tags.Insert(tag)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEntry):
			v.AddError("name", "a tag with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/tags/%d", tag.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"tag": tag}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTagHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r, "tag_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	tag, err := app.models.Tags.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name string `json:"name"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tag.Name = input.Name

	v := validator.New()

	if data.ValidateTag(v, tag); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Tags.Update(tag)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		case errors.Is(err, data.ErrDuplicateEntry):
			v.AddError("name", "a tag with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tag": tag}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTagHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r, "tag_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Tags.Delete(id)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "tag successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listPostTagsHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := app.readIDParam(r, "post_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "updated_at", "-id", "-name", "-updated_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	tags, metadata, err := app.models.Tags.GetAll(postID, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tags": tags, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
