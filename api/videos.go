package api

import (
	"net/http"
	"encoding/json"
	"github.com/mg4tv/kubrik/db"
	"github.com/mg4tv/kubrik/log"
	"github.com/jackc/pgx"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type videoResponse struct {
	Id             string `json:"id"`
	Title          string `json:"name"`
	OrganizationId string `json:"owner_id"`
}

type videoRequest struct {
	Id             *string `json:"id,omitempty"`
	Title          *string `json:"title,omitempty"`
	OrganizationId *string `json:"organization_id,omitempty"`
}

func listVideos(w http.ResponseWriter, _ *http.Request) {
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

func createVideo(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	var req videoRequest
	var err error
	var userId *string


	if userId, err = GetUserIdFromToken(r.Header.Get("authorization")); err != nil {
		if r.Header.Get("authorization") != "" {
			write403(w)
		} else {
			write401(w, &[]errorStruct{
				{
					Error:  "This endpoint requires a logged in user.",
					Fields: []string{"header: authorization"},
				},
			})
		}
		return
	}

	//FIXME: validate org
	/*if valid, vErrs := validateVideo(req, "create"); !valid {
		write422(w, vErrs)
		return
	}*/
	if err = decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	var isAuthorized bool
	if isAuthorized, err = IsAuthorized(*userId, *req.OrganizationId, "CREATE_VIDEO"); err != nil {
		write500(w)
		return
	} else if !isAuthorized {
		write403(w)
		return
	}


	newVideo, err := db.CreateVideo(*req.Title, *req.OrganizationId)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
		"err": err,
	}).Debug("Create Organization Failure")
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&videoResponse{
		Id:             newVideo.Id,
		Title:          newVideo.Title,
		OrganizationId: newVideo.OrganizationId,
	})
}

func showVideo(_ http.ResponseWriter, _ *http.Request) {
}

func RouteVideos(router *mux.Router) {
	sub := router.PathPrefix("/videos").Subrouter().StrictSlash(true)
	sub.HandleFunc("/", listVideos).Methods("GET")
	sub.Methods("GET").HandlerFunc(listVideos)
	sub.HandleFunc("/", createVideo).Methods("POST")
	sub.Methods("POST").HandlerFunc(createVideo)

	//router.DELETE("/videos/:id", deleteVideo)
	router.HandleFunc("/{id}", showVideo).Methods("GET")
	//router.PATCH("/videos/:id", partiallyUpdateVideo)
	//router.PUT("/videos/:id", updateVideo)
}
