package register

import (
	"fmt"
	"github.com/go-playground/validator"
)

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	Success bool   `json:"success"`
	Details string `json:"details"`
}

func (r *Request) Validate() error {
	newValidator := validator.New()
	if err := newValidator.Struct(r); err != nil {
		return fmt.Errorf("error validating: %w", err)
	}
	return nil
}
