package serializers

//easyjson:json
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
