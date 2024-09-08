package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/posts", app.listPostsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/posts/:post_id", app.showPostHandler)
	router.HandlerFunc(http.MethodPost, "/v1/posts", app.requireAuthorizedUser(app.createPostHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/posts/:post_id", app.requireAuthorizedUser(app.updatePostHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/posts/:post_id", app.requireAuthorizedUser(app.deletePostHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id/posts", app.listUserPostsHandler)

	router.HandlerFunc(http.MethodGet, "/v1/users", app.listUsersHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id", app.showUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.createUserHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/users/:user_id", app.requireAuthorizedUser(app.updateUserHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:user_id", app.requireAuthorizedUser(app.deleteUserHandler))

	router.HandlerFunc(http.MethodGet, "/v1/posts/:post_id/comments", app.listPostCommentsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/posts/:post_id/comments/:comment_id", app.showCommentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/posts/:post_id/comments", app.requireAuthorizedUser(app.createCommentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/posts/:post_id/comments/:comment_id", app.requireAuthorizedUser(app.updateCommentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/posts/:post_id/comments/:comment_id", app.requireAuthorizedUser(app.deleteCommentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id/comments", app.listUserCommentsHandler)

	router.HandlerFunc(http.MethodGet, "/v1/tags", app.listTagsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/tags/:tag_id", app.showTagHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tags", app.createTagHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/tags/:tag_id", app.updateTagHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/tags/:tag_id", app.deleteTagHandler)

	router.HandlerFunc(http.MethodGet, "/v1/posts/:post_id/tags", app.listPostTagsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/posts/:post_id/tags/:tag_id", app.addPostTagHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/posts/:post_id/tags/:tag_id", app.deletePostTagHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.logRequest(app.rateLimit(app.authenticate(router)))))
}
