package internal

import (
	"bytes"
	"encoding/base64"
	gor "github.com/askmi/gor/pkg/server"
	"log/slog"
	"net/http"
	"slices"
)

func UsernamePasswordAutenticator(u string) gor.Authenticator {
	return func(s gor.SecurityContext) (gor.SecurityContext, error) {
		b, ok := s.Identity().([]byte)
		if !ok {
			return gor.Rejected("no identity"), nil
		}
		user, p, ok := decodeBasic(b)
		if !ok {
			return gor.Rejected("invalid credentials"), nil
		}

		if string(user)+":"+string(p) != u {
			return gor.Rejected("invalid credentials"), nil
		}
		return gor.Authenticated(string(user), Principal{Username: string(user), Roles: []string{"user", "admin"}}), nil
	}
}

func decodeBasic(raw []byte) ([]byte, []byte, bool) {
	decoded, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return nil, nil, false
	}
	user, password, ok := bytes.Cut(decoded, []byte{':'})
	if !ok || len(password) == 0 {
		return nil, nil, false
	}
	return user, password, true
}

func Authorize(role string) gor.HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := gor.PrincipalFromContext[Principal](r.Context())
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !slices.Contains(p.Roles, role) {
				slog.InfoContext(r.Context(), p.Username+" does not have role: "+role)
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
