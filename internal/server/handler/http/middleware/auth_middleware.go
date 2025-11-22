package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func JWTAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "There is no JWT", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid JWT", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SignatureCheck(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inputHashStr := r.Header.Get("HashSHA256")
			if inputHashStr == "" {
				http.Error(w, "HashSHA256 header missing", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(body))
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			calcHash := mac.Sum(nil)
			inputHash, err := hex.DecodeString(inputHashStr)
			if err != nil {
				http.Error(w, "Invalid hash format", http.StatusBadRequest)
				return
			}
			if !hmac.Equal(calcHash, inputHash) {
				http.Error(w, "INVALID SIGNATURE", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
