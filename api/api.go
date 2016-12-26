package api

import (
	"net/http"
	"encoding/json"
)

type errorStruct struct {
	Error  string   `json:"error"`
	Fields []string `json:"fields"`
}

type errorResponse struct {
	HttpStatus int           `json:"httpStatus"`
	Message    string        `json:"message"`
	Errors     []errorStruct `json:"errors"`
}

func addContentTypeJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func write400(w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusBadRequest)
	encoder.Encode(errorResponse{
		HttpStatus: http.StatusBadRequest,
		Message: "Bad request",
		Errors: []errorStruct{},
	})
}

func write401(w http.ResponseWriter) {
}

func write403(w http.ResponseWriter) {
}

func write404(w http.ResponseWriter) {
}

func write415(w http.ResponseWriter) {
}

func write422(w http.ResponseWriter, errs []errorStruct) {
	encoder := json.NewEncoder(w)
	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusUnprocessableEntity)
	encoder.Encode(errorResponse{
		HttpStatus: http.StatusUnprocessableEntity,
		Message: "Unprocessable entity. JSON was parsed but did not match the expected structure",
		Errors: errs,
	})
}

func write500(w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	addContentTypeJSONHeader(w)
	w.WriteHeader(http.StatusInternalServerError)
	encoder.Encode(errorResponse{
		HttpStatus: http.StatusInternalServerError,
		Message: "Internal server error",
		Errors: []errorStruct{},
	})

}
