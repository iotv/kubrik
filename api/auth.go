package api

import (
	"net/http"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mg4tv/kubrik/db"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/mg4tv/kubrik/log"
	"github.com/mg4tv/kubrik/conf"
	"github.com/satori/go.uuid"
	"strings"
	"io/ioutil"
	"github.com/gorilla/mux"
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
	Code        *string `json:"code,omitempty"`
	ClientId    *string `json:"client_id,omitempty"`
	RedirectURI *string `json:"redirect_uri,omitemoty"`
}

type tokenRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

type jwtClaims struct {
	UserId *string `json:"uid,omitempty"`
	jwt.StandardClaims
}

// login is an httprouter.HandlerFunc which handles username/email & password login
// It can return the following HTTP statuses:
// 200 OK: The request was accepted and the body contains a signed JWT
// 400 Bad Request: The request was malformed and could not be parsed by JSON decoder
// 401 Unauthenticated: The credentials provided don't match a known user credential
// 422 Unprocessable Entity: The decoded JSON doesn't meet validation standards
// 500 Server Error:
func login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req tokenRequest
	var err error

	if err = decoder.Decode(&req); err != nil {
		// TODO: elaborate on the fields which are wrong
		write400(w)
		return
	}

	valid, eStructs := validateTokenRequest(&req)

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

	// TODO: handle connection errors and not just no rows
	if err != nil {
		// If the user is not found, we don't want to communicate that via a 404 error or timing on the 401 response.
		// Providing either of those would allow attackers to reduce their attack surface and only focus on existing users.
		// To prevent this, we manually set the user.EncryptedPassword so that the bcrypt hash comparison below will
		// Still happen and take time as if the user exists, but will always fail because []byte("FAKEPASSWORD") cannot
		// be the resulting bytes from our bcrypt hash as it is much too short.
		bcrypt.CompareHashAndPassword([]byte("FAKEPASSWORD"), []byte(*req.Password))
		write401(w, &[]errorStruct{
			{
				Error:  "Invalid Login/Password combination",
				Fields: []string{"password", "email", "username"},
			},
		})
		return
	}


	if err := bcrypt.CompareHashAndPassword(user.EncryptedPassword, []byte(*req.Password)); err != nil {
		write401(w, &[]errorStruct{
			{
				Error:  "Invalid Login/Password combination",
				Fields: []string{"password", "email", "username"},
			},
		})
		return
	}

	// TODO: check to make sure this config value exists... somehow
	testSigningKey := []byte(conf.Config.GetString("kubrik.secret"))
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
	// TODO: check that the client id is the one we have configured
	q.Add("client_id", *request.ClientId)
	q.Add("redirect_uri", *request.RedirectURI)
	q.Add("client_secret", conf.Config.GetString("facebook.client_secret"))
	q.Add("code", *request.Code)
	req.URL.RawQuery = q.Encode()

	log.Logger.WithField("URL", req.URL.RequestURI()).Info("our request")

	var fbResp serverFacebookTokenResponse
	resp, err := http.DefaultClient.Do(req)
	log.Logger.WithFields(logrus.Fields{
		"resp": resp,
	}).Debug("Convert code response")
	if resp.StatusCode != http.StatusOK {
		// FIXME: handle error within error inception
		body, _ := ioutil.ReadAll(resp.Body)
		log.Logger.
			WithField("URL", req.URL.RequestURI()).
			WithField("Body", body).
			Error("Facebook didn't like our request")
		return nil, errors.New("Bad Loginorino")
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fbResp); err != nil {
		return nil, err
	}

	if !validateFacebookTokenResponse(&fbResp) {
		return nil, errors.New("Bad login")
	}
	return &fbResp, nil
}

func loginOrSignUpWithFacebook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req clientFacebookTokenRequest
	var err error

	if err = decoder.Decode(&req); err != nil {
		log.Logger.Error("Failing to decode client Facebook Token Request in Login Or Sign Up With Facebook")
		write400(w)
		return
	}

	// Convert code into token
	// TODO: check validation
	validateFacebookTokenRequest(&req)
	accessToken, err := convertFacebookCodeToToken(req)
	if err != nil {
		log.Logger.Error("Failing to convert Facebook Code to token in Login or signup with facebook")
		write400(w)
		return
	}
	// TODO: check error and handle
	// Inspect token
	userAttrs, err := getFacebookUserAttributes(*accessToken.AccessToken)
	// TODO: validate fb token has fields we need and request didn't error
	log.Logger.WithFields(logrus.Fields{
		"email": *userAttrs.Email,
		"id":    *userAttrs.Id,
		"err":   err,
	}).Debug("Facebook user attribute response")

	var user *db.UserModel
	user, err = db.GetUserByFacebook(*userAttrs.Id)
	if err != nil {
		log.Logger.WithField("error", err).Debug("Get user by facebook error")
		user, err = db.CreateUserByFacebook(*userAttrs.Id, *userAttrs.Email)
		log.Logger.WithField("user", user).Debug("Create user response")
		if err != nil {
			log.Logger.WithField("error", err).Error("Failing to make new user")
			write400(w)
			return
		}
	}

	testSigningKey := []byte(conf.Config.GetString("kubrik.secret"))
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
	req.URL.RawQuery = q.Encode()

	var fbResp serverFacebookUserAttributes
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		// TODO: better error message
		return nil, errors.New("Bad Loginorino")
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fbResp); err != nil {
		return nil, err
	}
	return &fbResp, nil
}

func deauthFacebook(_ http.ResponseWriter, _ *http.Request) {
}

func convertGoogleToken(_ http.ResponseWriter, _ *http.Request) {
}

func jwtKeyFunc(_ *jwt.Token) (interface{}, error) {
	return []byte(conf.Config.GetString("kubrik.secret")), nil
}

func GetUserIdFromToken(header string) (*string, error) {
	//FIXME: sanity check that the user actually exists
	var jwtT *jwt.Token
	var err error
	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
		return nil, errors.New("Inmroper token form. Must begin with 'bearer '.")
	}
	if jwtT, err = jwt.ParseWithClaims(headerParts[1], &jwtClaims{}, jwtKeyFunc); err != nil {
		return nil, err
	}
	if claims, ok := jwtT.Claims.(*jwtClaims); ok && jwtT.Valid {
		if claims.UserId != nil {
			if _, err := uuid.FromString(*claims.UserId); err != nil {
				return nil, errors.New("Claimed user id is not a UUID")
			}
			return claims.UserId, nil
		}
		return nil, errors.New("Unspecified user id in claims")
	}
	return nil, errors.New("Invalid JWT")
}

func RouteAuth(router *mux.Router) {
	router.HandleFunc("/auth/login", login).Methods("POST")
	router.HandleFunc("/auth/facebook", loginOrSignUpWithFacebook).Methods("POST")
	router.HandleFunc("/auth/google", convertGoogleToken).Methods("POST")
	router.HandleFunc("/deauth/facebook", deauthFacebook).Methods("POST")
}
