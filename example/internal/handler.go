package internal

import (
	"context"
	"encoding/json"
	"errors"
	gor "gor/pkg/server"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"go.opentelemetry.io/otel/trace"
)

type H struct {
	s Service
}

func (h *H) GetUser(ctx context.Context, userID GetUserID) (GetUserResponse, error) {
	slog.InfoContext(ctx, "GetUser: ", "user_id", userID)

	id := int(userID)
	user, err := h.s.GetUser(ctx, id)
	if err != nil {
		return GetUserResponse{}, err
	}

	return GetUserResponse(user), nil
}

func (h *H) Me(ctx context.Context, p Principal) (Principal, error) {
	return p, nil
}

func (h *H) AddUser(ctx context.Context, req AddUserRequest) (AddUserResponse, error) {
	slog.InfoContext(ctx, "AddUser: ", "req", req)

	user, err := h.s.AddUser(ctx, AddUserRequestToUser(req))
	if err != nil {
		return AddUserResponse{}, err
	}

	return AddUserResponse(user), nil
}

func (h *H) EditUser(ctx context.Context, req EditUserRequest) (EditUserResponse, error) {
	slog.InfoContext(ctx, "EditUser: ", "req", req)
	user, err := h.s.EditUser(ctx, EditUserResponseToUser(req))
	if err != nil {
		return EditUserResponse{}, err
	}

	return EditUserResponse(user), nil
}

func (h *H) DeleteUser(ctx context.Context, req DeleteUserRequest) (Empty, error) {
	slog.InfoContext(ctx, "DeleteUser: ", "req", req)
	err := h.s.DeleteUser(ctx, req.ID)
	return nil, err
}

func (h *H) SearchUser(ctx context.Context, req SearchUserRequest) ([]GetUserResponse, error) {
	slog.InfoContext(ctx, "SearchUser:", "req", req)
	users, err := h.s.Search(ctx, req)
	if err != nil {
		return nil, err
	}
	return UserToGetUserResponse(users), nil
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "server: not found path "+r.RequestURI)
	http.NotFound(w, r)
}

func GetTrace(ctx context.Context, _ Empty) (map[string]string, error) {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		return nil, nil
	}

	return map[string]string{"trace_id": spanContext.TraceID().String()}, nil
}

func Hello(_ context.Context, name NameQuery) (string, error) {
	return "Hello world, " + string(name), nil
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), "websocket accept failed", "error", err)
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(1 << 20)

	ctx := context.WithoutCancel(r.Context())

	for {
		var message Message

		if err := wsjson.Read(ctx, c, &message); err != nil {
			return
		}

		response := Message{
			Type: "response",
			Text: "received: " + message.Text,
		}

		if err := wsjson.Write(ctx, c, response); err != nil {
			return
		}
	}
}

func AppErrorHandler(ctx context.Context, err error) gor.HTTPResponse {
	slog.ErrorContext(ctx, "server handle error", "error", err)
	statusCode := 500
	aType := "server_err"
	message := err.Error()
	switch {
	case errors.Is(err, ErrBadRequest):
		statusCode = 400
		aType = "bad_request"
		if cause := gor.Unwrap(err, 1); cause != nil {
			message = cause.Error()
		}
	case errors.Is(err, ErrNotFound):
		statusCode = 404
		aType = "not_found"
		if cause := gor.Unwrap(err, 1); cause != nil {
			message = cause.Error()
		}
	}
	m := map[string]string{
		"error":   aType,
		"message": message,
	}
	b, _ := json.Marshal(m)
	return gor.NewJSONResponse(statusCode, string(b))
}
