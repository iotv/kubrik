package api

import (
	"net/http"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/mg4tv/kubrik/db"
	"github.com/jackc/pgx"
	"github.com/mg4tv/kubrik/log"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type organizationResponse struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	OwnerId string `json:"owner_id"`
}

type organizationRequest struct {
	Name *string `json:"name,omitempty"`
}

func createOrganization(w http.ResponseWriter, r *http.Request) {
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
					Error:  "This endpoint requires a logged in user.",
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

	//FIXME: validate org
	/*if valid, vErrs := validateOrganization(req, "create"); !valid {
		write422(w, vErrs)
		return
	}*/
	newOrg, err := db.CreateOrganization(*req.Name, *userId, false)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Debug("Create Organization Failure")
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

func showOrganization(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	vars := mux.Vars(r)
	rawId, ok := vars["id"]

	if !ok {
		//TODO: debatable 500
		write400(w)
		return
	}
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
		Id:      org.Id,
		Name:    org.Name,
		OwnerId: org.OwnerId,
	})
}

func showOrganizationByName(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	vars := mux.Vars(r)
	name, ok := vars["name"]

	if !ok {
		//TODO: debatable 500
		write400(w)
		return
	}

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
		Id:      org.Id,
		Name:    org.Name,
		OwnerId: org.OwnerId,
	})
}

func IsAuthorized(userId, organizationId, permission string) (bool, error) {
	return true, nil
}

func RouteOrganization(router *mux.Router) {
	sub := router.PathPrefix("/organizations").Subrouter().StrictSlash(true)

	//router.GET("/organizations", listOrganizations)
	sub.Methods("POST").HandlerFunc(createOrganization)
	sub.HandleFunc("/", createOrganization).Methods("POST")

	//router.DELETE("/organizations/:id", deleteOrganization)
	sub.HandleFunc("/{id}", showOrganization).Methods("GET")
	//router.PATCH("/organizations/:id", partiallyUpdateOrganization)
	//router.PUT("/organizations/:id", updateOrganization)

	router.HandleFunc("/organizationsByName/{name}", showOrganizationByName).Methods("GET")
}
