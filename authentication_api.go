package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

const SERVER_ADDRESS = "localhost:8080"
const TOKEN_EXPIRATION = time.Hour * 24
const TOKEN_ISSUER = "authentication-api"
const SIGNING_KEY = "AllYourBase"
const REDIS_ADDRESS = "localhost:6379"
const REDIS_PASSWORD = ""

type UserDto struct {
	Username string
	Password string
}

type UserDao interface {
	Get(username string) (string, error)
	Set(username string, password string) error
}

type UserDaoImpl struct {
	ctx context.Context
	rdb *redis.Client
}

func (u UserDaoImpl) Get(username string) (string, error) {
	return u.rdb.Get(u.ctx, username).Result()
}

func (u UserDaoImpl) Set(username string, password string) error {
	return u.rdb.SetNX(u.ctx, username, password, 0).Err()
}

func getUserDto(body io.Reader) (UserDto, error) {
	var user UserDto

	b, errB := io.ReadAll(body)
	if errB != nil {
		return user, errB
	}

	errU := json.Unmarshal(b, &user)
	if errU != nil {
		return user, errU
	}

	return user, nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getRegisterHandler(u UserDao) func(_ http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, errU := getUserDto(r.Body)
		if errU != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, errU.Error())
			return
		}

		err := u.Set(user.Username, user.Password)
		if err != nil {
			print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getLoginHandler(u UserDao, now func() time.Time) func(_ http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, errU := getUserDto(r.Body)
		if errU != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, errU.Error())
			return
		}

		p, err := u.Get(user.Username)
		if err != nil || p != user.Password {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims := &jwt.StandardClaims{
			ExpiresAt: now().Add(TOKEN_EXPIRATION).Unix(),
			Issuer:    TOKEN_ISSUER,
			Subject:   user.Username,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		ss, err := token.SignedString([]byte(SIGNING_KEY))
		if err != nil {
			print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		io.WriteString(w, ss)
	}
}

func setupMux(serveMux *http.ServeMux, u UserDao, now func() time.Time) {
	serveMux.HandleFunc("/", healthCheckHandler)
	serveMux.HandleFunc("/auth/register", getRegisterHandler(u))
	serveMux.HandleFunc("/auth/login", getLoginHandler(u, now))
}

func main() {
	u := UserDaoImpl{
		ctx: context.Background(),
		rdb: redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDRESS,
			Password: REDIS_PASSWORD,
			DB:       0,
		}),
	}
	setupMux(http.DefaultServeMux, u, time.Now)

	fmt.Print("starting authentication-api on http://", SERVER_ADDRESS, "\n")
	log.Fatal(http.ListenAndServe(SERVER_ADDRESS, nil))
}
