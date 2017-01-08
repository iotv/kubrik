package api

func validateTokenRequest(r *tokenRequest) (bool, *[]errorStruct) {

	valid := true
	var eStructs []errorStruct
	if r.Username == nil && r.Email == nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have either an email or username",
			Fields: []string{"username", "email"},
		})
	} else if r.Username != nil && r.Email != nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have only an email or username, not both",
			Fields: []string{"username", "email"},
		})
	}

	if r.Password == nil {
		valid = false
		eStructs = append(eStructs, errorStruct{
			Error:  "Request must have a password",
			Fields: []string{"password"},
		})
	}

	if !valid {
		return false, &eStructs
	}

	return true, nil
}

func validateFacebookTokenRequest(r *clientFacebookTokenRequest) bool {
	return true
}

func validateFacebookTokenResponse(r *serverFacebookTokenResponse) bool {
	if r.AccessToken == nil {
		return false
	}

	return true
}