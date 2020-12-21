package auth

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const tokenExpires = 60 * time.Second

type MemStore struct {
	users    map[string]User
	sessions map[string]string
	token    *TokenFactory
}

//func NewMemStore(signingKey string) (*MemStore, error) {
func NewMemStore(signingKey string) (AuthManager, error) {
	m := &MemStore{
		users:    make(map[string]User),
		sessions: make(map[string]string),
		token:    NewTokenFactory([]byte(signingKey), tokenExpires),
	}
	return m, nil
}

func (s *MemStore) CreateUser(name, pass string) (User, error) {
	// TODO: check for name collisions

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "failed to generate pass")
	}
	u := User{
		id:   newID(time.Now()),
		Name: name,
		Hash: hash,
		//LastLogin: time.Unix(0,0),
	}
	s.users[u.ID()] = u
	return u, nil
}

func (s *MemStore) GetUser(id string) (User, bool) {
	u, found := s.users[id]
	return u, found
}

func (s *MemStore) UpdateUser(user User) error {
	_, found := s.GetUser(user.ID())
	if !found {
		return errors.New("invalid user")
	}
	s.users[user.ID()] = user
	return nil
}

func (s *MemStore) DeleteUser(user User) {
	delete(s.users, user.ID())
}

////////////////////////////////////////////////////////////////////////////////

func (s *MemStore) CreateSession(name, pass string) (string, error) {
	user := User{}
	for _, u := range s.users {
		if u.Name == name {
			user = u
			break
		}
	}

	if err := bcrypt.CompareHashAndPassword(user.Hash, []byte(pass)); err != nil {
		return "", errors.Wrap(err, "invalid password")
	}

	user.LastLogin = time.Now()
	if err := s.UpdateUser(user); err != nil {
		return "", errors.Wrap(err, "failed to update user.LastLogin")
	}

	token := s.token.Generate()
	s.sessions[token] = user.ID()
	return token, nil
}

func (s *MemStore) GetSession(token string) (User, bool) {
	id, found := s.sessions[token]
	if !found {
		return User{}, false
	}
	u, found := s.users[id]
	return u, found
}

func (s *MemStore) UpdateSession(token string) (string, error) {
	user, found := s.GetSession(token)
	if !found {
		return "", errors.New("invalid session token")
	}
	newToken := s.token.Generate()
	s.sessions[newToken] = user.ID()
	return newToken, nil
}

func (s *MemStore) DeleteSession(token string) {
	delete(s.sessions, token)
}

////////////////////////////////////////////////////////////////////////////////

func (s *MemStore) WriteSessionHeader(w http.ResponseWriter, token string) {
	s.token.WriteSessionHeader(w, token)
}

func (s *MemStore) WriteSessionCookie(w http.ResponseWriter, token string) {
	s.token.WriteSessionCookie(w, token)
}

func (s *MemStore) GetSessionToken(r *http.Request) (string, error) {
	return s.token.GetSessionToken(r)
}

////////////////////////////////////////////////////////////////////////////////

const (
	base36      string = "0123456789abcdefghijklmnopqrstuvwxyz"
	base36Len   uint64 = uint64(len(base36))
	idBytesTime int    = 9
	idBytesRand int    = 7
)

func newID(ts time.Time) string {
	buf := make([]byte, idBytesTime+idBytesRand)

	t := uint64(ts.UnixNano() / int64(time.Millisecond))
	for i := idBytesTime - 1; i >= 0; i-- {
		buf[i] = base36[t%base36Len]
		t /= base36Len
	}

	_, err := rand.Read(buf[idBytesTime:])
	if err != nil {
		// Pretty screwed if you can't use the system RNG
		panic(errors.Wrap(err, "system failure for crypto/rand"))
	}
	for i := idBytesTime; i < idBytesTime+idBytesRand; i++ {
		buf[i] = base36[uint64(buf[i])%base36Len]
	}

	return string(buf)
}
