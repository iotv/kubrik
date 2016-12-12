package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

type pwLoginRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type errorStruct struct {
	Error  string    `json:"error"`
	Fields *[]string `json:"fields,omitempty"`
}

type errorResponse struct {
	HttpStatus int           `json:"httpStatus"`
	Message    string        `json:"message"`
	Errors     []errorStruct `json:"errors"`
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req pwLoginRequest

	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; utf-8")
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(&errorResponse{
			HttpStatus: 400,
			Message:    "Bad request",
			Errors: []errorStruct{
				{
					Error:  "Malformed JSON",
					Fields: &[]string{"body"},
				},
			},
		})
		return
	}
	fmt.Fprintf(w, "Token")
}

func main() {
	router := httprouter.New()
	router.POST("/users/login", login)
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
