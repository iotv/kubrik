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

type organizationRequest struct {
	Name *string `json:"name,omitempty"`
}

func createOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	var req organizationRequest
	var err error
	var userId *string

	if userId, err = GetUserIdFromToken(r.Header.Get("authorization")); err != nil {
		if r.Header.Get("authorization") != "" {
			write403(w)
		} else {
			write401(w, &[]errorStruct{
				{
					Error: "This endpoint requires a logged in user.",
					Fields: []string{"header: authorization"},
				},
			})
		}
		return
	}

	if err = decoder.Decode(&req); err != nil {
		write400(w)
		return
	}

	/*if valid, vErrs := validateOrganization(req, "create"); !valid {
		write422(w, vErrs)
		return
	}*/
	newOrg, err := db.CreateOrganization(db.OrganizationModel{
		Name: *req.Name,
		OwnerId: *userId,
	})
	if err != nil {
		write500(w)
		return
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&organizationResponse{
		Id:      newOrg.Id,
		Name:    newOrg.Name,
		OwnerId: newOrg.OwnerId,
	})
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
	router.POST("/organizations", createOrganization)

	//router.DELETE("/organizations/:id", deleteOrganization)
	router.GET("/organizations/:id", showOrganization)
	//router.PATCH("/organizations/:id", partiallyUpdateOrganization)
	//router.PUT("/organizations/:id", updateOrganization)

	router.GET("/organizationsByName/:name", showOrganizationByName)
}
