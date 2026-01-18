package serializers

//easyjson:json
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
