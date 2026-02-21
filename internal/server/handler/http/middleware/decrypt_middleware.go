package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/pkg/crypto"
)

func RequestDecrypt(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			contentType := r.Header.Get("Content-Type")
			if contentType != "application/octet-stream" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			decryptedData, err := crypto.Decrypt(privateKey, body)
			if err != nil {
				http.Error(w, "failed to decrypt body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		})
	}
}
