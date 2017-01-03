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

type debugFacebookTokenResponse struct {
	Data *struct {
		AppId       *string               `json:"app_id,omitempty"`
		Application *string               `json:"application,omitempty"`

		Error *struct {
			Code    *int    `json:"code,omitempty"`
			Message *string `json:"message,omitempty"`
			Subcode *int    `json:"subcode,omitempty"`
		}                                 `json:"error,omitempty"`

		ExpiresAt *int                    `json:"expires_at,omitempty"`
		IsValid   *bool                   `json:"is_valid,omitempty"`
		IssuedAt  *int                    `json:"issued_at,omitempty"`
		Metadata  *map[string]interface{} `json:"metadata,omitempty"`
		ProfileId *string                 `json:"profile_id,omitempty"`
		Scopes    *[]string               `json:"scopes,omitempty"`
		UserId    *string                 `json:"user_id,omitempty"`
	} `json:"data,omitempty"`
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
	fbToken, err := debugFacebookToken(*accessToken.AccessToken)
	// TODO: validate fb token has fields we need and request didn't error

	var user *db.UserModel
	user, err = db.GetUserByFacebook(*fbToken.Data.UserId)
	if err != pgx.ErrNoRows {
		// TODO create user
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

func debugFacebookToken(accessToken string) (*debugFacebookTokenResponse, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/debug_token", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("input_token", accessToken)
	// TODO add access token here

	var fbResp debugFacebookTokenResponse
	resp, err := http.DefaultClient.Do(req)
	// TODO: Check status code for error
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
