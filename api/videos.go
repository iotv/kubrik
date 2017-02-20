package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"github.com/mg4tv/kubrik/db"
	"github.com/jackc/pgx"
)

func listVideos(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	encoder := json.NewEncoder(w)

	videos, err := db.ListVideos(20)
	if err == pgx.ErrNoRows {
		write404(w)
		return
	} else if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&videos)
}

func createVideo(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func showVideo(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
}

func RouteVideos(router *httprouter.Router) {
	router.GET("/videos", listVideos)
	router.POST("/videos", createVideo)

	//router.DELETE("/videos/:id", deleteVideo)
	router.GET("/videos/:id", showVideo)
	//router.PATCH("/videos/:id", partiallyUpdateVideo)
	//router.PUT("/videos/:id", updateVideo)
}
