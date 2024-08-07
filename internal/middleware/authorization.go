package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models"
	"github.com/RIBorisov/gophermart/internal/service"
)

type Auth struct {
	Service *service.Service
}

func CheckAuth(svc *service.Service) *Auth {
	return &Auth{Service: svc}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const accessDenied = "Access Denied"
		rCtx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.Service.Log.Err(accessDenied, "Authorization header has not provided")
			http.Error(w, accessDenied, http.StatusUnauthorized)
			return
		}

		userID := getUserID(authHeader[7:], a.Service.Config.Secret.SecretKey, a.Service.Log)
		if userID == "" {
			a.Service.Log.Err(accessDenied, "Authorization header contains no userID")
			http.Error(w, accessDenied, http.StatusUnauthorized)
			return
		}

		newCtx := context.WithValue(rCtx, models.CtxUserIDKey, userID)
		rWithCtx := r.WithContext(newCtx)
		next.ServeHTTP(w, rWithCtx)
	})
}

func getUserID(tokenString, secretKey string, log *logger.Log) string {
	claims := &service.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		log.Err("failed parse with claims tokenString: ", err)
		return ""
	}
	if !token.Valid {
		log.Err("Token is not valid: ", token)
		return ""
	}

	return claims.UserID
}
