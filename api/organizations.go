package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/mg4tv/kubrik/db"
	"github.com/jackc/pgx"
)

type organizationResponse struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	OwnerId string `json:"owner_id"`
}

func showOrganization(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	encoder := json.NewEncoder(w)
	rawId := p.ByName("id")
	if _, err := uuid.FromString(rawId); err != nil {
		write400(w)
		return
	}

	org, err := db.GetOrganizationById(rawId)
	if err == pgx.ErrNoRows {
		write404(w)
		return
	} else if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&organizationResponse{
		Id:         org.Id,
		Name:       org.Name,
		OwnerId:    org.OwnerId,
	})
}

func showOrganizationByName(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	r.Header.Get("authorization")
	encoder := json.NewEncoder(w)
	name := p.ByName("name")

	org, err := db.GetOrganizationByName(name)
	if err == pgx.ErrNoRows {
		write404(w)
		return
	} else if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&organizationResponse{
		Id:         org.Id,
		Name:       org.Name,
		OwnerId:    org.OwnerId,
	})
}

func RouteOrganization(router *httprouter.Router) {
	//router.GET("/organizations", listOrganizations)
	//router.POST("/organizations", createOrganization)

	//router.DELETE("/organizations/:id", deleteOrganization)
	router.GET("/organizationss/:id", showOrganization)
	//router.PATCH("/organizations/:id", partiallyUpdateOrganization)
	//router.PUT("/organizations/:id", updateOrganization)

	router.GET("/organizationsByName/:name", showOrganizationByName)
}
