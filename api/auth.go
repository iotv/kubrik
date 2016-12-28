package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/mg4tv/kubrik/db"
)

type tokenRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type tokenResponse struct {
	Token  string `json:"token"`
	Header string `json:"header"`
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req tokenRequest
	var err error

	if err = decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	valid := true
	var eStructs []errorStruct
	if req.Username == nil && req.Email == nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have either an email or username",
			Fields: []string{"username", "email"},
		})
	} else if req.Username != nil && req.Email != nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have only an email or username, not both",
			Fields: []string{"username", "email"},
		})
	}

	if req.Password == nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have a password",
			Fields: []string{"password"},
		})
	}

	if !valid {
		write422(w, eStructs)
		return
	}

	var user *db.UserModel
	if req.Username != nil {
		user, err = db.GetUserByUsername(*req.Username)
	} else {
		user, err = db.GetUserByEmail(*req.Email)
	}

	if err != nil {
		write404(w)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.EncryptedPassword, []byte(*req.Password)); err != nil {
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
