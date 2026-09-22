package internal

import (
	"encoding/json"
	"errors"
	gor "gor/pkg/server"
	"io"
	"net/http"
	"strconv"
	"time"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

type (
	Empty     = *struct{}
	NameQuery string
	GetUserID int

	Principal struct {
		Username string
		Roles    []string
	}

	AddUserRequest struct {
		Name  string
		Email string
	}
	EditUserRequest struct {
		ID    int
		Name  string
		Email string
	}
	DeleteUserRequest struct {
		ID int
	}

	GetUserResponse struct {
		ID       int
		Name     string
		Email    string
		CreateAt time.Time
	}

	SearchUserRequest struct {
		Email string
	}

	AddUserResponse struct {
		ID       int
		Name     string
		Email    string
		CreateAt time.Time
	}

	EditUserResponse struct {
		ID       int
		Name     string
		Email    string
		CreateAt time.Time
	}

	Message struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
)

func (q *NameQuery) DecodeFromHTTPRequest(req *http.Request) error {
	value := req.URL.Query().Get("name")
	if value == "" {
		return errors.New("name parameter is missing")
	}
	*q = NameQuery(value)
	return nil
}

func (t *SearchUserRequest) DecodeFromHTTPRequest(req *http.Request) error {
	value := req.URL.Query().Get("email")
	t.Email = value
	return nil
}

func (t *DeleteUserRequest) DecodeFromHTTPRequest(r *http.Request) error {
	ID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}
	t.ID = ID
	return nil
}

func (p *Principal) DecodeFromHTTPRequest(r *http.Request) error {
	principal, ok := gor.PrincipalFromContext[Principal](r.Context())
	if !ok {
		p.Username = "anonymous"
	} else {
		*p = principal
	}
	return nil
}

func (t *GetUserID) DecodeFromHTTPRequest(r *http.Request) error {
	v := r.PathValue("id")
	if v == "" {
		return errors.Join(ErrBadRequest, errors.New("userID required"))
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return errors.Join(ErrBadRequest, errors.New("userID must be number"))

	}
	if n < 0 {
		return errors.Join(ErrBadRequest, errors.New("userID can not be negative"))
	}

	*t = GetUserID(n)
	return nil
}

func (t *AddUserRequest) DecodeFromHTTPRequest(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, t)
}

func (t *EditUserRequest) DecodeFromHTTPRequest(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, t)
}

/***********************************************************
   MAPPERS
***********************************************************/

func UserToGetUserResponse(users []User) []GetUserResponse {
	res := make([]GetUserResponse, 0, len(users))
	for i, u := range users {
		resp := GetUserResponse{}
		resp.ID = u.ID
		resp.Name = u.Name
		resp.Email = u.Email
		resp.CreateAt = u.CreateAt
		res[i] = resp
	}

	return res
}

func AddUserRequestToUser(r AddUserRequest) User {
	user := User{}
	user.Name = r.Name
	user.Email = r.Email
	return user
}

func EditUserResponseToUser(r EditUserRequest) User {
	user := User{}
	user.ID = r.ID
	user.Name = r.Name
	user.Email = r.Email
	return user
}
