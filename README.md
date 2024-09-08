# Blogging API

This is a RESTful API for a blogging platform that allows users to create, read, update, and delete blog posts, comments, users, and tags. It also includes user authentication and authorization features. The API is built using the Go programming language and the `httprouter` package for routing.

## Features

- **User Management**: Create, update, delete, and view users.
- **Post Management**: Create, update, delete, and view blog posts.
- **Comment Management**: Create, update, delete, and view comments on posts.
- **Tag Management**: Create, update, delete, and view tags.
- **Post-Tag Association**: Add and remove tags from posts.
- **Authentication**: Secure endpoints using JSON Web Token (JWT) authentication.
- **Authorization**: Restrict access to certain actions for authenticated users only.
- **Rate Limiting**: Limit the number of API requests to avoid abuse.
- **Error Handling**: Custom responses for 404 Not Found and 405 Method Not Allowed.
- **Healthcheck**: Endpoint to verify the health of the API.

## Endpoints

### Healthcheck

- `GET /v1/healthcheck`: Check if the API is running.

### Users

- `GET /v1/users`: List all users.
- `GET /v1/users/:user_id`: Retrieve a specific user.
- `POST /v1/users`: Create a new user.
- `PATCH /v1/users/:user_id`: Update a user's information (requires authentication).
- `DELETE /v1/users/:user_id`: Delete a user (requires authentication).
- `GET /v1/users/:user_id/posts`: List all posts from a specific user.
- `GET /v1/users/:user_id/comments`: List all comments made by a specific user.

### Posts

- `GET /v1/posts`: List all posts.
- `GET /v1/posts/:post_id`: Retrieve a specific post.
- `POST /v1/posts`: Create a new post (requires authentication).
- `PATCH /v1/posts/:post_id`: Update an existing post (requires authentication).
- `DELETE /v1/posts/:post_id`: Delete a post (requires authentication).
- `GET /v1/posts/:post_id/comments`: List all comments on a specific post.
- `GET /v1/posts/:post_id/tags`: List all tags associated with a specific post.

### Comments

- `GET /v1/posts/:post_id/comments`: List all comments on a post.
- `GET /v1/posts/:post_id/comments/:comment_id`: Retrieve a specific comment.
- `POST /v1/posts/:post_id/comments`: Create a new comment on a post (requires authentication).
- `PATCH /v1/posts/:post_id/comments/:comment_id`: Update an existing comment (requires authentication).
- `DELETE /v1/posts/:post_id/comments/:comment_id`: Delete a comment (requires authentication).

### Tags

- `GET /v1/tags`: List all tags.
- `GET /v1/tags/:tag_id`: Retrieve a specific tag.
- `POST /v1/tags`: Create a new tag.
- `PATCH /v1/tags/:tag_id`: Update an existing tag.
- `DELETE /v1/tags/:tag_id`: Delete a tag.
- `POST /v1/posts/:post_id/tags/:tag_id`: Add a tag to a post.
- `DELETE /v1/posts/:post_id/tags/:tag_id`: Remove a tag from a post.

### Authentication

- `POST /v1/tokens/authentication`: Generate a JWT token for user authentication.

## Middleware

The API includes several middleware functions to handle common tasks:

- **Recover Panic**: Gracefully handle unexpected panics and prevent the application from crashing.
- **Log Request**: Log incoming requests for debugging and monitoring purposes.
- **Rate Limiting**: Limit the number of requests a client can make within a certain time frame.
- **Authentication**: Authenticate users using JWT tokens.
- **Authorization**: Restrict access to certain routes for authenticated users only.

## Setup

1. Install Go: Ensure that Go is installed on your machine.
2. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/blogging-api.git
   ```
3. Navigate to the project directory:
   ```bash
   cd blogging-api
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Run the server:
   ```bash
   go run main.go
   ```

## Configuration

- The API requires a `.env` file or configuration management for settings like database connections and JWT secret keys.
- Ensure that the environment variables are set for running the server in production.

## Authentication

Authentication is handled using JWT tokens. To access secured routes, users must include a valid JWT token in the `Authorization` header in the format:

```
Authorization: Bearer <token>
```

Tokens are generated via the `/v1/tokens/authentication` endpoint after a user successfully logs in.

## Error Handling

Custom error responses are provided for:

- 404 Not Found
- 405 Method Not Allowed
- 401 Unauthorized for unauthorized access attempts
