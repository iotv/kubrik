package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var v interface{}

	if err := decoder.Decode(&v); err != nil {
		fmt.Fprintf(w, "Error")
		return
	}

	switch v := v.(type) {
	case map[string]interface{}:
		_, oku := v["username"]
		_, okp := v["password"]
		if oku && okp {
			fmt.Fprintf(w, "tokenz")
		} else {
			fmt.Fprintf(w, "Missing keys")
		}
	default:
		fmt.Fprintf(w, "Unrecognized")
	}
}

func main() {
	router := httprouter.New()
	router.POST("/users/login", login)
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
