package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rs/cors"
	"github.com/julienschmidt/httprouter"
	"github.com/mg4tv/kubrik/api"
	"github.com/urfave/negroni"
)

var Serve = &cobra.Command{
	Use:   "serve",
	Short: "run a web server",
	Run:   serve,
}

var port string
var host string

func init() {
	RootCmd.AddCommand(Serve)
	Serve.Flags().StringVarP(&host, "host", "H", "", "The host to serve on")
	Serve.Flags().StringVarP(&port, "port", "p", "4000", "The port HTTP server will run on")
	viper.BindPFlag("host", Serve.Flags().Lookup("host"))
	viper.BindPFlag("port", Serve.Flags().Lookup("port"))
}

func serve(cmd *cobra.Command, args []string) {
	corsMiddleware := cors.Default()
	router := httprouter.New()
	api.RouteAuth(router)
	api.RouteUser(router)
	n := negroni.Classic()
	n.Use(corsMiddleware)
	n.UseHandler(router)
	n.Run(host + ":" + port)
}
