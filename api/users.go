package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"github.com/mg4tv/kubrik/db"
	"github.com/jackc/pgx"
	"github.com/satori/go.uuid"
)

type userResponse struct {
	Id       string  `json:"id"`
	Username *string `json:"username,omitempty"`
	Email    string  `json:"email"`
}

type userRequest struct {
	Id                   *string `json:"id,omitempty"`
	Username             *string `json:"username,omitempty"`
	Email                *string `json:"email,omitempty"`
	Password             *string `json:"password,omitempty"`
	PasswordConfirmation *string `json:"password_confirmation,omitempty"`
}

// validateUser ensures that a user request is valid.
// If the request is valid, nil is returned, otherwise a populated
// errorStruct is returned, identifying the errors encountered during
// validation
func validateUser(u userRequest, act string) (bool, *[]errorStruct) {
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
				"password_confirmation",
			},
		})
	}

	if (u.Password == nil && u.PasswordConfirmation != nil) || (u.Password != nil && u.PasswordConfirmation == nil) {
		valid = false
	}

	if !valid {
		return false, &vErrs
	}

	return true, nil
}

// createUser is an httprouter handler function which responds to POST requests for users
// It performs the CRUD create operation in a RESTful manner by validating the request
// and writing the user to the database
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

	if valid, vErrs := validateUser(req, "create"); !valid {
		write422(w, vErrs)
		return
	}
	// TODO: encrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 10)
	if err != nil {
		write500(w)
		return
	}

	newUser, err := db.CreateUser(db.UserModel{
		Username:          req.Username,
		Email:             *req.Email,
		EncryptedPassword: hash,
	})
	if err != nil {
		if pgErr := err.(pgx.PgError); pgErr.Code == "23505" /*duplicate key violates unique constraint*/ {
			write409(w, &[]errorStruct{
				{
					Error: pgErr.ConstraintName + "must be unique",
					Fields: []string{
						pgErr.ConstraintName,
					},
				},
			})
			return
		}
		write500(w)
		return
	}

	// TODO: extract for update/partial update
	// FIXME: do this in a transaction?
	// Check if username is taken by groups first
	if req.Username != nil {
		if _, err := db.GetOrganizationByName(*req.Username); err == pgx.ErrNoRows {
			if _, err := db.CreateOrganization(*req.Username, newUser.Id, true); err != nil {
				if pgErr := err.(pgx.PgError); pgErr.Code == "23505" /*duplicate key violates unique constraint*/ {
					db.DeleteUser(newUser.Id) // FIXME: handle error
					write409(w, &[]errorStruct{
						{
							Error: pgErr.ConstraintName + "must be unique",
							Fields: []string{
								pgErr.ConstraintName,
							},
						},
					})
					return
				} else {
					db.DeleteUser(newUser.Id) // FIXME: handle error
					write500(w)
					return
				}
			}
		} else if err == nil {
			write409(w, &[]errorStruct{
				{
					Error: "Username must be unique",
					Fields: []string{
						"username",
					},
				},
			})
			return
		} else {
			write500(w)
			return
		}
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
	rawId := p.ByName("id")
	if _, err := uuid.FromString(rawId); err != nil {
		write400(w)
		return
	}

	user, err := db.GetUserById(rawId)
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
	username := p.ByName("username")

	user, err := db.GetUserByUsername(username)
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

	router.GET("/userByUsername/:username", showUserByUsername)
	//router.GET("/usersByEmail/:email", showUserByEmail)
}
