package register

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	Success bool   `json:"success"`
	Details string `json:"details"`
}
