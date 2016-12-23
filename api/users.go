package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/mg4tv/kubrik/db"
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

func validateUser(u userRequest, act string) *errorStruct {
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
		//FIXME: send 400
		return
	}

	if vErrs := validateUser(req, "create"); vErrs != nil {
		//FIXME: send 422
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), 10) // TODO: send 500 if fail
	newUser, err := db.CreateUser(db.UserModel{
		Username:          *req.Username,
		Email:             *req.Email,
		EncryptedPassword: hash,
	})
	if err != nil {
		//FIXME: send 500
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&userResponse{
		Id:       newUser.Id,
		Username: newUser.Username,
		Email:    newUser.Email,
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func listUsers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func partiallyUpdateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func showUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func updateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func RouteUser(router *httprouter.Router) {
	router.GET("/users", listUsers)
	router.POST("/users", createUser)

	router.DELETE("/users/:id", deleteUser)
	router.GET("/users/:id", showUser)
	router.PATCH("/users/:id", partiallyUpdateUser)
	router.PUT("/users/:id", updateUser)
}
