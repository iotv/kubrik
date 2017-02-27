package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rs/cors"
	"github.com/mg4tv/kubrik/api"
	"github.com/urfave/negroni"
	"github.com/meatballhat/negroni-logrus"
	"github.com/mg4tv/kubrik/log"
	"github.com/gorilla/mux"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "run a web server",
	Run:   serve,
}

var port string
var host string
var secret string

func init() {
	RootCmd.AddCommand(ServeCmd)
	ServeCmd.Flags().StringVarP(&host, "host", "H", "", "The host to serve on")
	ServeCmd.Flags().StringVarP(&port, "port", "p", "4000", "The port HTTP server will run on")
	ServeCmd.Flags().StringVarP(&secret, "secret", "s", "", "Facebook Secret")
	viper.BindPFlag("host", ServeCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", ServeCmd.Flags().Lookup("port"))
	viper.BindPFlag("secret", ServeCmd.Flags().Lookup("secret"))
}

func serve(cmd *cobra.Command, args []string) {
	corsMiddleware := cors.Default()
	router := mux.NewRouter()

	// Initiate subroutes
	api.RouteAuth(router)
	api.RouteOrganization(router)
	api.RouteUser(router)
	api.RouteVideos(router)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negronilogrus.NewMiddlewareFromLogger(log.Logger, "HTTP"))
	n.Use(corsMiddleware)
	n.UseHandler(router)
	n.Run(host + ":" + port)
}
