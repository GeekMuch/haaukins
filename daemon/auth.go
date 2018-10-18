package daemon

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aau-network-security/go-ntp/store"
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	USERNAME_KEY    = "un"
	SUPERUSER_KEY   = "su"
	VALID_UNTIL_KEY = "vu"
)

var (
	InvalidUsernameOrPassErr = errors.New("Invalid username or password")
	InvalidTokenFormatErr    = errors.New("Invalid token format")
	TokenExpiredErr          = errors.New("Token has expired")
	UnknownUserErr           = errors.New("Unknown user")
	EmptyUserErr             = errors.New("Username cannot be empty")
	EmptyPasswdErr           = errors.New("Password cannot be empty")
)

type Authenticator interface {
	TokenForUser(username, password string) (string, error)
	AuthenticateUserByToken(context.Context, string) (context.Context, error)
}

type auth struct {
	us  store.UserStore
	key string
}

type su struct{}
type us struct{}

func NewAuthenticator(us store.UserStore, key string) Authenticator {
	return &auth{
		us:  us,
		key: key,
	}
}

func (a *auth) TokenForUser(username, password string) (string, error) {
	username = strings.ToLower(username)

	if username == "" {
		return "", EmptyUserErr
	}

	if password == "" {
		return "", EmptyPasswdErr
	}

	u, err := a.us.GetUserByUsername(username)
	if err != nil {
		return "", InvalidUsernameOrPassErr
	}

	if ok := u.IsCorrectPassword(password); !ok {
		return "", InvalidUsernameOrPassErr
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		USERNAME_KEY:    u.Username,
		SUPERUSER_KEY:   u.SuperUser,
		VALID_UNTIL_KEY: time.Now().Add(31 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(a.key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *auth) AuthenticateUserByToken(ctx context.Context, t string) (context.Context, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(a.key), nil
	})
	if err != nil {
		return ctx, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims[USERNAME_KEY].(string)
		if !ok {
			return ctx, InvalidTokenFormatErr
		}

		u, err := a.us.GetUserByUsername(username)
		if err != nil {
			return ctx, UnknownUserErr
		}

		validUntil, ok := claims[VALID_UNTIL_KEY].(float64)
		if !ok {
			return ctx, InvalidTokenFormatErr
		}

		if int64(validUntil) < time.Now().Unix() {
			return ctx, TokenExpiredErr
		}

		ctx = context.WithValue(ctx, su{}, u.SuperUser)
		ctx = context.WithValue(ctx, us{}, u.Username)

		return ctx, nil
	}

	return ctx, err
}
