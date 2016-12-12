package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/negroni"
)

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func main() {
	n := negroni.Classic()
	n.UseHandlerFunc(handle)
	n.Run(":8080")
}
