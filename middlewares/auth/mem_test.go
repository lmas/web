package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %q, expected nil", err)
	}
}

func TestUseCaseProcedure(t *testing.T) {
	store, err := NewMemStore("super secret signing key")
	if err != nil {
		t.Fatalf("failed to create mem store: %v", err)
	}
	name, pass, token := "tester", "secret password", ""
	var user User

	t.Run("create user", func(t *testing.T) {
		user, err = store.CreateUser(name, pass)
		assertNoError(t, err)
		if user.IsZero() {
			t.Errorf("got user with no ID")
		}
		if user.Name != name {
			t.Errorf("got user.Name %w, expected %q", user.Name, name)
		}
		err = bcrypt.CompareHashAndPassword(user.Hash, []byte(pass))
		assertNoError(t, err)
	})

	t.Run("get user", func(t *testing.T) {
		u, found := store.GetUser(user.ID())
		if !found || u.IsZero() {
			t.Errorf("found no user")
		}
		if u.ID() != user.ID() {
			t.Errorf("got user with ID %q, expected user with ID %q", u.ID(), user.ID())
		}
		if !u.LastLogin.IsZero() {
			t.Errorf("got user with LastLogin %q, expected zero", u.LastLogin)
		}
	})

	t.Run("login user and create new session", func(t *testing.T) {
		token, err = store.CreateSession(name, pass)
		assertNoError(t, err)
		if token == "" {
			t.Errorf("got empty token")
		}
	})

	t.Run("get user from session", func(t *testing.T) {
		u, found := store.GetSession(token)
		if !found || u.IsZero() {
			t.Errorf("found no user")
		}
		if u.ID() != user.ID() {
			t.Errorf("got user with ID %q, expected user with ID %q", u.ID(), user.ID())
		}
		if u.LastLogin.IsZero() {
			t.Errorf("got user with no LastLogin")
		}
	})

	t.Run("update user", func(t *testing.T) {
		user.Name = "new tester"
		err := store.UpdateUser(user)
		assertNoError(t, err)
		u, found := store.GetUser(user.ID())
		if !found || u.IsZero() {
			t.Errorf("found no user")
		}
		if u.Name != user.Name {
			t.Errorf("got user with Name %q, expected user with Name %q", u.Name, user.Name)
		}
	})

	t.Run("refresh session and update token", func(t *testing.T) {
		newToken, err := store.UpdateSession(token)
		assertNoError(t, err)
		if newToken == "" || newToken == token {
			t.Errorf("got old token, expected a new one")
		}
		token = newToken
	})

	t.Run("logout session", func(t *testing.T) {
		store.DeleteSession(token)
		_, found := store.GetSession(token)
		if found {
			t.Errorf("got user, expected none")
		}
	})

	t.Run("delete user", func(t *testing.T) {
		store.DeleteUser(user)
		u, found := store.GetUser(user.ID())
		if found && !u.IsZero() {
			t.Errorf("found user, expected none")
		}
	})
}
