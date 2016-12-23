package main

import (
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/satori/go.uuid"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/bcrypt"
	"github.com/mg4tv/kubrik/api"
)

var pgPool *pgx.ConnPool

type userResponse struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type errorStruct struct {
	Error  string   `json:"error"`
	Fields []string `json:"fields"`
}

type errorResponse struct {
	HttpStatus int           `json:"httpStatus"`
	Message    string        `json:"message"`
	Errors     []errorStruct `json:"errors"`
}

type tokenResponse struct {
	Token  string `json:"token"`
	Header string `json:"header"`
}

func listUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	encoder := json.NewEncoder(w)

	pgConn, _ := pgPool.Acquire()
	defer pgPool.Release(pgConn)

	users := []userResponse{}
	rows, _ := pgConn.Query("select * from users")
	for rows.Next() {
		var id string
		var username string
		var email string
		var encrypted_password []byte
		rows.Scan(&id, &username, &email, &encrypted_password)
		users = append(users, userResponse{
			Username: username,
			Email:    email,
		})
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&users)
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req map[string]interface{}

	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(&errorResponse{
			HttpStatus: http.StatusBadRequest,
			Message:    "Bad request",
			Errors: []errorStruct{
				{
					Error:  "Malformed JSON",
					Fields: []string{"body"},
				},
			},
		})
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		encoder.Encode(&errorResponse{
			HttpStatus: http.StatusUnprocessableEntity,
			Message:    "Unprocessable entity",
			Errors:     eStructs,
		})
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

func main() {
	pgPool, _ = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "db",
			User:     "postgres",
			Password: "postgres",
			Database: "mg4",
		},
	})
	corsMiddleware := cors.Default()
	router := httprouter.New()
	router.POST("/auth/login", login)
	api.RouteUser(router)
	n := negroni.Classic()
	n.Use(corsMiddleware)
	n.UseHandler(router)
	n.Run(":8080")
}
