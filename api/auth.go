package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mg4tv/kubrik/db"
	"github.com/spf13/viper"
	"github.com/jackc/pgx"
)

type serverFacebookTokenResponse struct {
	AccessToken *string `json:"access_token,omitempty"`
	TokenType   *string `json:"token_type,omitempty"`
	ExpiresIn   *int    `json:"expires_in,omitempty"`
}

type serverFacebookUserAttributes struct {
	Id    *string `json:"id,omitempty"`
	Email *string `json:"email,omitempty"`
}

type clientFacebookTokenRequest struct {
	Code        *string `json:"code"`
	ClientId    *string `json:"client_id"`
	RedirectURI *string `json:"redirect_uri"`
}

type tokenRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
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

	// TODO: move secret to config
	testSigningKey := []byte("secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"uid": user.Id,
	})
	tokenString, _ := token.SignedString(testSigningKey)

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&tokenResponse{
		Token:     string(tokenString),
		TokenType: "bearer",
	})
}

func convertFacebookCodeToToken(request clientFacebookTokenRequest) (*serverFacebookTokenResponse, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/oauth/access_token", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("client_id", *request.ClientId)
	q.Add("redirect_uri", *request.RedirectURI)
	q.Add("client_secret", "")
	q.Add("code", *request.Code)

	var fbResp serverFacebookTokenResponse
	resp, err := http.DefaultClient.Do(req)
	// TODO: Check status code for error
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fbResp); err != nil {
		return nil, err
	}
	return &fbResp, nil
}

func validateFacebookTokenRequest(request clientFacebookTokenRequest) error {
	return nil
}

func loginOrSignUpWithFacebook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req clientFacebookTokenRequest
	var err error

	if err = decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	// Convert code into token
	validateFacebookTokenRequest(req)
	accessToken, err := convertFacebookCodeToToken(req)
	// TODO: check error and handle
	// Inspect token
	userAttrs, err := getFacebookUserAttributes(*accessToken.AccessToken)
	// TODO: validate fb token has fields we need and request didn't error

	var user *db.UserModel
	user, err = db.GetUserByFacebook(*userAttrs.Id)
	if err != pgx.ErrNoRows {
		user, err = db.CreateUserByFacebook(*userAttrs.Id, *userAttrs.Email)
		if err != nil {
			// TODO: handle
		}
	}

	testSigningKey := []byte(viper.GetString("kubrik.secret"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"uid": user.Id,
	})
	tokenString, _ := token.SignedString(testSigningKey)

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&tokenResponse{
		Token:     string(tokenString),
		TokenType: "bearer",
	})

}

func getFacebookUserAttributes(accessToken string) (*serverFacebookUserAttributes, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/me", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("fields", "id,email")
	q.Add("access_token", accessToken)

	var fbResp serverFacebookUserAttributes
	resp, err := http.DefaultClient.Do(req)
	// TODO: Check response code
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fbResp); err != nil {
		return nil, err
	}
	return &fbResp, nil
}

func deauthFacebook(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func convertGoogleToken(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func RouteAuth(router *httprouter.Router) {
	router.POST("/auth/login", login)
	router.POST("/auth/facebook", loginOrSignUpWithFacebook)
	router.POST("/auth/google", convertGoogleToken)
	router.POST("/deauth/facebook", deauthFacebook)
}
