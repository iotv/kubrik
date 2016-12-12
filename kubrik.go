package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	router := httprouter.New()
	router.POST("/users/login", login)
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
