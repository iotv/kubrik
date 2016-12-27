package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/mg4tv/kubrik/db"
	"github.com/satori/go.uuid"
	"github.com/jackc/pgx"
)

type userResponse struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type userRequest struct {
	Id                   *string `json:"id"`
	Username             *string `json:"username"`
	Email                *string `json:"email"`
	Password             *string `json:"password"`
	PasswordConfirmation *string `json:"passwordConfirmation"`
}

func validateUser(u userRequest, act string) *[]errorStruct {
	vErrs := []errorStruct{}
	valid := true
	if u.Id == nil && act != "create" {
		valid = false
	}

	if u.Username == nil {
		valid = false
		vErrs = append(vErrs, errorStruct{
			Error: "Username cannot be empty",
			Fields: []string{
				"username",
			},
		})
	}

	if u.Email == nil {
		valid = false
		vErrs = append(vErrs, errorStruct{
			Error: "Email cannot be empty",
			Fields: []string{
				"email",
			},
		})
	}

	if u.Password == nil && u.PasswordConfirmation == nil && act == "create" {
		valid = false
		vErrs = append(vErrs, errorStruct{
			Error: "Password and password confirmation cannot be empty",
			Fields: []string{
				"password",
				"passwordConfirmation",
			},
		})
	}

	if (u.Password == nil && u.PasswordConfirmation != nil) || (u.Password != nil && u.PasswordConfirmation == nil) {
		valid = false
	}

	if !valid {
		return nil
	}

	return nil
}

func createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO: pull these from a pool
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req userRequest
	var err error

	if err = decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	if vErrs := validateUser(req, "create"); vErrs != nil {
		write422(w, *vErrs)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 10)
	if err != nil {
		write500(w)
		return
	}
	newUser, err := db.CreateUser(db.UserModel{
		Username:          *req.Username,
		Email:             *req.Email,
		EncryptedPassword: hash,
	})
	if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&userResponse{
		Id:       newUser.Id,
		Username: newUser.Username,
		Email:    newUser.Email,
	})
}

func deleteUser(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func listUsers(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func partiallyUpdateUser(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func showUser(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	encoder := json.NewEncoder(w)
	rawID := p.ByName("id")
	if _, err := uuid.FromString(rawID); err != nil {
		write400(w)
		return
	}

	user, err := db.GetUserById(rawID)
	if err == pgx.ErrNoRows {
		write404(w)
		return
	} else if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&userResponse{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	})
}

func showUserByUsername(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	encoder := json.NewEncoder(w)
	rawUsername := p.ByName("username")
	if _, err := uuid.FromString(rawUsername); err != nil {
		write400(w)
		return
	}

	user, err := db.GetUserByUsername(rawUsername)
	if err == pgx.ErrNoRows {
		write404(w)
		return
	} else if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&userResponse{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	})
}

func updateUser(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func RouteUser(router *httprouter.Router) {
	router.GET("/users", listUsers)
	router.POST("/users", createUser)

	router.DELETE("/users/:id", deleteUser)
	router.GET("/users/:id", showUser)
	router.PATCH("/users/:id", partiallyUpdateUser)
	router.PUT("/users/:id", updateUser)

	router.GET("/users/byUsername/:username", showUserByUsername)
}
