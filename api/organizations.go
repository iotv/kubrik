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

type groupResponse struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []permissionResponse `json:"permissions"`
}

type permissionResponse struct {
	Id     string `json:"id"`
	TypeId string `json:"type_id"`
	Name   string `json:"name"`
}

type organizationResponse struct {
	Id      string          `json:"id"`
	Name    string          `json:"name"`
	OwnerId string          `json:"owner_id"`
	Groups  []groupResponse `json:"groups"`
}

type groupRequest struct {
	Name        *string   `json:"name,omitempty"`
	IsPublic    *bool     `json:"is_public,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
}

type organizationRequest struct {
	Name    *string       `json:"name,omitempty"`
	OwnerId *string       `json:"name,omitempty"`
	Groups  *groupRequest `json:"groups,omitempty"`
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

	resp := organizationResponse{
		Id:      org.Id,
		Name:    org.Name,
		OwnerId: org.OwnerId,
		Groups:  []groupResponse{},
	}

	for _, group := range org.Groups {
		gResp := groupResponse{
			Id:          group.Id,
			Name:        group.Name,
			Permissions: []permissionResponse{},
		}
		for _, permission := range group.Permissions {
			gResp.Permissions = append(gResp.Permissions, permissionResponse{
				Id:     permission.Id,
				TypeId: permission.PermissionTypeId,
				Name:   permission.PermissionTypeName,
			})
		}
		resp.Groups = append(resp.Groups, gResp)
	}

	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(&resp)
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
	orgRouter := router.PathPrefix("/organizations").Subrouter().StrictSlash(true)

	// Root paths
	//orgRouter.Methods("GET").HandlerFunc(listOrganizations)
	//orgRouter.HandleFunc("/", listOrganizations).Methods("GET")
	orgRouter.Methods("POST").HandlerFunc(createOrganization)
	orgRouter.HandleFunc("/", createOrganization).Methods("POST")

	// By Id Paths
	//orgRouter.HandleFunc("/{id}", deleteOrganization).Methods("DELETE")
	orgRouter.HandleFunc("/{id}", showOrganization).Methods("GET")
	//orgRouter.HandlerFunc("/{id}", partiallyUpdateOrganization).Methods("PATCH")
	//orgRouter.HandlerFunc("/{id}", updateOrganization).Methods("PUT")

	// By Name Paths

	// Groups subroutes
	//orgRouter.HandleFunc("/{orgId}/groups", listOrganizationGroups).Methods("GET")
	//orgRouter.HandleFunc("/{orgId}/groups", createOrganizationGroup).Methods("POST")
	//orgRouter.HandleFunc("/{orgId}/groups/{groupId}", showOrganizationGroup).Methods("GET")
	//orgRouter.HandleFunc("/{orgId}/groups/{groupId}", deleteOrganizationGroup).Methods("DELETE")
	//orgRouter.HandleFunc("/{orgId}/groups/{groupId}", partiallyUpdateOrganizationGroup).Methods("PATCH")
	//orgRouter.HandleFunc("/{orgId}/groups/{groupId}", updateOrganizationGroup).Methods("PUT")

	router.HandleFunc("/orgsByName/{name}", showOrganizationByName).Methods("GET")
}
