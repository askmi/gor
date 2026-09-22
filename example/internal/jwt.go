package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	gor "gor/pkg/server"
	"strings"
	"time"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidBearer  = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidValue   = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidHeader  = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidPayload = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidSig     = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidSub     = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidIss     = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidExp     = fmt.Errorf("%w", ErrInvalidToken)
	ErrTokenExpired   = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidAlg     = fmt.Errorf("%w", ErrInvalidToken)
	ErrInvalidKID     = fmt.Errorf("%w", ErrInvalidToken)
)

type (
	jwtSecurityContext struct {
		claims    map[string]any
		expiredAt time.Time
		alg       string
		kid       string
		sub       string
		iss       string
		raw       []byte
		sig       []byte
	}
)

func (jwt *jwtSecurityContext) IsAuthenticated() bool  { return false }
func (jwt *jwtSecurityContext) IdentityString() string { return jwt.sub }
func (jwt *jwtSecurityContext) Identity() any          { return jwt.claims }

var JwtAuthenticator = func(s gor.SecurityContext) (gor.SecurityContext, error) {
	b, ok := s.Identity().([]byte)
	if !ok {
		return gor.Rejected("invalid bearer"), nil
	}
	jwt, err := DecodeBearer(b)
	if err != nil {
		return gor.Rejected("invalid jwt"), nil
	}

	err = jwt.Validate()
	if err != nil {
		return gor.Rejected("invalid jwt"), nil
	}

	return gor.Authenticated(jwt.sub, jwt), nil
}

func (jwt *jwtSecurityContext) Validate() error {
	if err := jwt.ValidateIssuer("iss"); err != nil {
		return err
	}
	if err := jwt.ValidateExpiration(); err != nil {
		return err
	}
	if err := jwt.ValidateSignature(); err != nil {
		return err
	}

	return nil
}

func (jwt *jwtSecurityContext) ValidateIssuer(iss string) error {
	if iss == jwt.iss {
		return ErrInvalidIss
	}

	return nil
}

func (jwt *jwtSecurityContext) ValidateExpiration() error {
	if jwt.expiredAt.IsZero() {
		return ErrInvalidExp
	}
	if !time.Now().Before(jwt.expiredAt) {
		return ErrTokenExpired
	}

	return nil
}

func (jwt *jwtSecurityContext) ValidateSignature() error {
	if len(jwt.sig) == 0 {
		return ErrInvalidSig
	}
	if jwt.kid == "" {
		return ErrInvalidKID
	}

	return nil
}

func DecodeBearer(raw []byte) (jwtSecurityContext, error) {
	var jwt jwtSecurityContext
	if len(raw) == 0 {
		return jwt, ErrInvalidValue
	}

	parts := strings.Split(string(string(raw)), ".")
	if len(parts) != 3 {
		return jwt, ErrInvalidValue
	}

	headerEncoded, payloadEncoded, sigEncoded := []byte(parts[0]), []byte(parts[1]), []byte(parts[2])

	claims := map[string]any{}
	dst := make([]byte, max(len(headerEncoded), max(len(payloadEncoded), len(sigEncoded))))

	n, err := base64.StdEncoding.Decode(dst, headerEncoded)
	if err != nil || n == 0 {
		return jwt, ErrInvalidHeader
	}
	if err := json.Unmarshal(dst[:n], &claims); err != nil {
		return jwt, ErrInvalidHeader
	}

	n, err = base64.RawURLEncoding.Decode(dst, payloadEncoded)
	if err != nil || n == 0 {
		return jwt, ErrInvalidPayload
	}
	if err := json.Unmarshal(dst[:n], &claims); err != nil {
		return jwt, ErrInvalidPayload
	}

	n, err = base64.RawURLEncoding.Decode(dst, sigEncoded)
	if err != nil || n == 0 {
		return jwt, ErrInvalidSig
	}

	sig := make([]byte, n)
	copy(sig, dst[:n])

	alg, ok := claims["alg"].(string)
	if !ok {
		return jwt, ErrInvalidAlg
	}

	// kid, ok := claims["kid"].(string)
	// if !ok {
	// return nil, ErrInvalidKID
	// }

	sub, ok := claims["sub"].(string)
	if !ok {
		return jwt, ErrInvalidSub
	}

	// iss, ok := claims["iss"].(string)
	// if !ok {
	// return nil, ErrInvalidIss
	// }

	unixSec, ok := claims["iat"].(float64)
	if !ok {
		return jwt, ErrInvalidExp
	}
	exp := time.Unix(int64(unixSec), 0)

	jwt.kid = "kid"
	jwt.iss = "iss"
	jwt.raw = raw
	jwt.alg = alg
	jwt.claims = claims
	jwt.sub = sub
	jwt.sig = sig
	jwt.expiredAt = exp

	return jwt, nil
}
