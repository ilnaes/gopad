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

// func Login(w http.ResponseWriter, r *http.Request) {
// 	var creds Credentials
// 	// Get the JSON body and decode into credentials
// 	err := json.NewDecoder(r.Body).Decode(&creds)
// 	if err != nil {
// 		// If the structure of the body is wrong, return an HTTP error
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	// Get the expected password from our in memory map
// 	expectedPassword, ok := users[creds.Username]

// 	// If a password exists for the given user
// 	// AND, if it is the same as the password we received, the we can move ahead
// 	// if NOT, then we return an "Unauthorized" status
// 	if !ok || expectedPassword != creds.Password {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}

// 	// Declare the expiration time of the token
// 	// here, we have kept it as 5 minutes
// 	expirationTime := time.Now().Add(5 * time.Minute)
// 	// Create the JWT claims, which includes the username and expiry time
// 	claims := &Claims{
// 		Uid: creds.Username,
// 		StandardClaims: jwt.StandardClaims{
// 			// In JWT, the expiry time is expressed as unix milliseconds
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	// Declare the token with the algorithm used for signing, and the claims
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	if godotenv.Load() != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// 	jwtKey := os.Getenv("JWT_SECRET")

// 	// Create the JWT string
// 	tokenString, err := token.SignedString(jwtKey)
// 	if err != nil {
// 		// If there is an error in creating the JWT return an internal server error
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	// Finally, we set the client cookie for "token" as the JWT we just generated
// 	// we also set an expiry time which is the same as the token itself
// 	http.SetCookie(w, &http.Cookie{
// 		Name:    "token",
// 		Value:   tokenString,
// 		Expires: expirationTime,
// 	})
// }
