package auth

import "time"

type User struct {
	id        string // protected
	Name      string
	Hash      []byte
	LastLogin time.Time
}

func (u User) ID() string {
	return u.id
}

func (u User) IsZero() bool {
	return u.ID() == ""
}

type AuthManager interface {
	CreateUser(string, string) (User, error)
	GetUser(string) (User, bool)
	UpdateUser(User) error
	DeleteUser(User)

	CreateSession(string, string) (string, error)
	GetSession(string) (User, bool)
	UpdateSession(string) (string, error)
	DeleteSession(string)
}
