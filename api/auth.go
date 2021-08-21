package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt"
)

func login(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var requestBody struct {
		ID       int64  `json:"id"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		writeError(w, "Error decoding body", http.StatusBadRequest)
		return
	}

	// Get teacher record for requested id
	teacher, err := getTeacherByID(requestBody.ID)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Password match?
	err = bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(requestBody.Password))
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Err was nil so passwords matched

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role":       "teacher",
		"teacher_id": teacher.ID,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(hmacSampleSecret))

	var response struct {
		Token string `json:"token"`
		Error bool   `json:"error"`
	}
	response.Token = tokenString
	response.Error = false

	// respond to the client
	responseBytes, _ := json.Marshal(response)
	w.Write(responseBytes)

}

//
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate credentials
		err := validateJWT(r.Header.Get("authorization"))
		if err != nil {
			log.Println(err)

			writeError(w, "TokenInvalid", http.StatusUnauthorized)
			return
		}

		// Just assume teacher creds for now, but really this would parse the JWT
		// so that we can know the role & ID of the logged in user
		ctx := context.WithValue(r.Context(), "role", "teacher")

		fmt.Println("Auth OK")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateJWT(tokenString string) error {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSampleSecret), nil
	})

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("InvalidToken")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("Claims Teacher ID: ", claims["teacher_id"])
	} else {
		return fmt.Errorf("InvalidToken")
	}

	return nil
}
