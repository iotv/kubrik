package cmd

import (
	"github.com/spf13/cobra"
	"github.com/rs/cors"
	"github.com/julienschmidt/httprouter"
	"github.com/mg4tv/kubrik/api"
	"github.com/urfave/negroni"
)

func init() {
	RootCmd.AddCommand(Serve)
}

var Serve = &cobra.Command{
	Use: "serve",
	Short: "run a web server",
	Run: func(cmd *cobra.Command, args []string) {
		corsMiddleware := cors.Default()
		router := httprouter.New()
		api.RouteAuth(router)
		api.RouteUser(router)
		n := negroni.Classic()
		n.Use(corsMiddleware)
		n.UseHandler(router)
		n.Run(":8080")
	},
}