package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Uid string `json:"uid"`
	jwt.StandardClaims
}

// userid -> token, err
func (s *Server) signJWT(claim Claims) (string, error) {
	claim.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(s.secret)
}

// token -> userid, ok
func (s *Server) parseJWT(token string) (string, bool) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})

	if err != nil {
		return "", false
	}

	if claim, ok := parsedToken.Claims.(*Claims); ok && parsedToken.Valid {
		return claim.Uid, true
	}

	return "", false
}

func (s *Server) middleware(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		extractedToken := strings.Split(token, "Bearer ")

		if len(extractedToken) != 2 {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		uid, ok := s.parseJWT(extractedToken[1])

		if ok {
			ctx := context.WithValue(r.Context(), "Uid", uid)
			next(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token", http.StatusForbidden)
		}
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var user Credentials
	if json.Unmarshal(reqBody, &user) != nil {
		http.Error(w, "Bad format", http.StatusForbidden)
	}

	filter := bson.D{{"username", user.Username}, {"password", user.Password}}

	res := s.users.FindOne(context.TODO(), filter)

	if res.Err() == nil {
		token, _ := s.signJWT(Claims{
			Uid: user.Username,
		})
		fmt.Fprintf(w, token)
	} else {
		http.Error(w, "Invalid credentials", http.StatusForbidden)
	}
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var user Credentials
	if json.Unmarshal(reqBody, &user) != nil {
		http.Error(w, "Bad format", http.StatusForbidden)
	}

	filter := bson.D{{"username", user.Username}}
	update := bson.D{{"$setOnInsert", bson.D{{"password", user.Password}}}}
	opts := options.Update().SetUpsert(true)

	res, err := s.users.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}

	if res.UpsertedCount != 0 {
		token, _ := s.signJWT(Claims{
			Uid: user.Username,
		})
		fmt.Fprintf(w, token)
	} else {
		http.Error(w, "Already exists", http.StatusForbidden)
	}
}
