// Package internal — OpenAPI specification for the user endpoints.
//
// @title       GoR Example User API
// @version     1.0
// @description Typed user endpoints served by the GoR example service.
// @description Every handler is a plain Go function of the form
// @description `func(context.Context, Req) (Resp, error)`; this file adds the
// @description OpenAPI surface without touching those functions.
// @BasePath    /api/v1
// @Host        localhost:8080
// @Schemes     http
//
// @tag.name        User
// @tag.description Create, read, update, delete and search users.
//
// @tag.name        Identity
// @tag.description The caller's own authenticated identity.
//
// @securityDefinitions.basic BasicAuth
// @securityDefinitions.basic.description Standard HTTP Basic Authentication. The
// @securityDefinitions.basic.description example service accepts `admin:admin`.
package internal

import "context"

// UserAPI is the documented user surface. Every method mirrors a handler on H,
// so the compile-time assertions below fail the build whenever a handler's
// signature drifts away from the annotations that describe it.
type UserAPI interface {
	AddUser(ctx context.Context, req AddUserRequest) (AddUserResponse, error)
	GetUser(ctx context.Context, userID GetUserID) (GetUserResponse, error)
	EditUser(ctx context.Context, req EditUserRequest) (EditUserResponse, error)
	DeleteUser(ctx context.Context, req DeleteUserRequest) (Empty, error)
	SearchUser(ctx context.Context, req SearchUserRequest) ([]GetUserResponse, error)
	Me(ctx context.Context, p Principal) (Principal, error)
}

// Both the real handler and the documented wrapper must satisfy the same
// interface: the spec cannot describe an endpoint the service does not serve,
// and a changed handler signature breaks the build here rather than silently
// shipping a stale spec.
var (
	_ UserAPI = (*H)(nil)
	_ UserAPI = (*UserAPISpec)(nil)
)

// UserAPISpec carries the OpenAPI annotations. It delegates every call, so it adds
// documentation and nothing else — wrap a handler with it, or leave it out of
// the router entirely, without changing behaviour.
type UserAPISpec struct {
	api UserAPI
}

func NewUserAPISpec(api UserAPI) *UserAPISpec {
	return &UserAPISpec{api: api}
}

// AddUser godoc
//
//	@OperationID	AddUser
//	@Summary		Create a user
//	@Description	Registers a new user and returns the stored representation.
//	@Description	The server owns the identity and creation time: `ID` is assigned
//	@Description	on insert and `CreatedAt` is set from the database clock, so both
//	@Description	fields are ignored if supplied by the caller.
//	@Description	Requires the `admin` role.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Param			request	body		AddUserRequest	true	"User to create"
//	@Success		201		{object}	AddUserResponse	"User created"
//	@Failure		400		{object}	ErrorResponse	"Malformed body"
//	@Failure		401		{object}	ErrorResponse	"Missing or invalid credentials"
//	@Failure		403		{object}	ErrorResponse	"Caller lacks the admin role"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/user [post]
func (s *UserAPISpec) AddUser(ctx context.Context, req AddUserRequest) (AddUserResponse, error) {
	return s.api.AddUser(ctx, req)
}

// GetUser godoc
//
//	@OperationID	GetUser
//	@Summary		Get a user by ID
//	@Description	Returns the user identified by the path value.
//	@Description	The identifier is decoded and validated before the handler runs:
//	@Description	a missing, non-numeric or negative `id` is rejected as `400` and
//	@Description	never reaches the business logic.
//	@Tags			User
//	@Produce		json
//	@Security		BasicAuth
//	@Param			id	path		int				true	"User identifier"	minimum(0)
//	@Success		200	{object}	GetUserResponse	"User found"
//	@Failure		400	{object}	ErrorResponse	"Identifier missing, not a number, or negative"
//	@Failure		401	{object}	ErrorResponse	"Missing or invalid credentials"
//	@Failure		404	{object}	ErrorResponse	"No user with that identifier"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/user/{id} [get]
func (s *UserAPISpec) GetUser(ctx context.Context, userID GetUserID) (GetUserResponse, error) {
	return s.api.GetUser(ctx, userID)
}

// EditUser godoc
//
//	@OperationID	EditUser
//	@Summary		Update a user
//	@Description	Replaces the mutable fields of an existing user and returns the
//	@Description	stored representation. The target is taken from the body's `ID`.
//	@Description	`CreatedAt` is immutable and is not read from the request.
//	@Description	Requires the `admin` role.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Param			request	body		EditUserRequest		true	"User to update"
//	@Success		200		{object}	EditUserResponse	"User updated"
//	@Failure		400		{object}	ErrorResponse		"Malformed body"
//	@Failure		401		{object}	ErrorResponse		"Missing or invalid credentials"
//	@Failure		403		{object}	ErrorResponse		"Caller lacks the admin role"
//	@Failure		404		{object}	ErrorResponse		"No user with that identifier"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/user [put]
func (s *UserAPISpec) EditUser(ctx context.Context, req EditUserRequest) (EditUserResponse, error) {
	return s.api.EditUser(ctx, req)
}

// DeleteUser godoc
//
//	@OperationID	DeleteUser
//	@Summary		Delete a user
//	@Description	Removes the user identified by the path value. The operation is
//	@Description	idempotent: deleting an absent user is not reported as an error.
//	@Description	Requires the `admin` role.
//	@Tags			User
//	@Produce		json
//	@Security		BasicAuth
//	@Param			id	path	int	true	"User identifier"	minimum(0)
//	@Success		204	"User deleted; no content returned"
//	@Failure		400	{object}	ErrorResponse	"Identifier missing or not a number"
//	@Failure		401	{object}	ErrorResponse	"Missing or invalid credentials"
//	@Failure		403	{object}	ErrorResponse	"Caller lacks the admin role"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/user/{id} [delete]
func (s *UserAPISpec) DeleteUser(ctx context.Context, req DeleteUserRequest) (Empty, error) {
	return s.api.DeleteUser(ctx, req)
}

// SearchUser godoc
//
//	@OperationID	SearchUser
//	@Summary		Search users
//	@Description	Returns the users matching the supplied criteria. An absent
//	@Description	`email` matches every user, so a bare request lists them all.
//	@Tags			User
//	@Produce		json
//	@Security		BasicAuth
//	@Param			email	query		string			false	"Filter by e-mail address"
//	@Success		200		{array}		GetUserResponse	"Matching users; an empty array when nothing matched"
//	@Failure		401		{object}	ErrorResponse	"Missing or invalid credentials"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/user [get]
func (s *UserAPISpec) SearchUser(ctx context.Context, req SearchUserRequest) ([]GetUserResponse, error) {
	return s.api.SearchUser(ctx, req)
}

// Me godoc
//
//	@OperationID	GetMe
//	@Summary		Get the authenticated identity
//	@Description	Returns the principal built by the application's authenticator,
//	@Description	including the roles that route-scoped authorization reads.
//	@Description	Unauthenticated callers are reported as the `anonymous` username
//	@Description	rather than rejected, which makes this endpoint a convenient way
//	@Description	to inspect how a credential was interpreted.
//	@Tags			Identity
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	Principal		"The caller's principal"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/user/me [get]
func (s *UserAPISpec) Me(ctx context.Context, p Principal) (Principal, error) {
	return s.api.Me(ctx, p)
}

// ErrorResponse is the failure body produced by AppErrorHandler. It exists so
// the generated spec can name a schema for every documented error.
type ErrorResponse struct {
	// Error is a stable, machine-readable class such as "bad_request",
	// "not_found" or "server_err".
	Error string `json:"error" example:"not_found"`
	// Message is the human-readable cause.
	Message string `json:"message" example:"Get: User not found"`
}
