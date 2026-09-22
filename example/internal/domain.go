package internal

import "time"

type (
	User struct {
		ID        int
		Name      string
		Email     string
		CreatedAt time.Time
	}
)

var ZeroUser = User{}
