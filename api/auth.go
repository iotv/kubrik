package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
)

type tokenResponse struct {
	Token  string `json:"token"`
	Header string `json:"header"`
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req map[string]interface{}

	if err := decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	valid := true
	var eStructs []errorStruct
	_, uexist := req["username"]
	_, eexist := req["email"]
	if !(uexist || eexist) {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have either an email or username",
			Fields: []string{"username", "email"},
		})
	}
	p, pexist := req["password"]
	if !pexist {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have a password",
			Fields: []string{"password"},
		})
	}
	password, pIsString := p.(string)
	if pexist && !pIsString {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Password value must be a JSON string",
			Fields: []string{"password"},
		})
	}

	if !valid {
		write422(w, eStructs)
		return
	}

	// Compare to itself. There's no database to check
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		encoder.Encode(&errorResponse{
			HttpStatus: http.StatusUnauthorized,
			Message:    "Unauthorized",
			Errors: []errorStruct{
				{
					Error:  "Invalid Login/Password combination",
					Fields: []string{"password", "email", "username"},
				},
			},
		})
		return
	}

	testSigningKey := []byte("secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"uid": uuid.NewV4(),
	})
	tokenString, _ := token.SignedString(testSigningKey)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder.Encode(&tokenResponse{
		Token:  string(tokenString),
		Header: "Bearer: " + string(tokenString),
	})
}

func convertFacebookToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func convertGoogleToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func RouteAuth(router *httprouter.Router) {
	router.POST("/auth/login", login)
	router.POST("/auth/fromFacebook", convertFacebookToken)
	router.POST("/auth/fromGoogle", convertGoogleToken)
}
