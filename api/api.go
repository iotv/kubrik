package api

type errorStruct struct {
	Error  string   `json:"error"`
	Fields []string `json:"fields"`
}