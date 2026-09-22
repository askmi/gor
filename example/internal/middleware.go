package internal

import (
	gor "gor/pkg/server"
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
		user, p, ok := gor.DecodeBasic(b)
		if !ok {
			return gor.Rejected("invalid credentials"), nil
		}

		if string(user)+":"+string(p) != u {
			return gor.Rejected("invalid credentials"), nil
		}
		return gor.Authenticated(string(user), Principal{Username: string(user), Roles: []string{"user", "admin"}}), nil
	}
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
