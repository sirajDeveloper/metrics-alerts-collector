package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type hashWriter struct {
	http.ResponseWriter
	buf        []byte
	secret     string
	status     int
	headerSent bool
}

func (hw *hashWriter) Write(b []byte) (int, error) {
	if !hw.headerSent {
		hw.status = http.StatusOK
	}
	hw.buf = append(hw.buf, b...)
	return len(b), nil
}

func (hw *hashWriter) WriteHeader(code int) {
	if !hw.headerSent {
		hw.status = code
	}
}

func (hw *hashWriter) writeHashedResponse() {
	if hw.headerSent {
		return
	}
	hw.headerSent = true

	if hw.secret != "" && len(hw.buf) > 0 {
		mac := hmac.New(sha256.New, []byte(hw.secret))
		mac.Write(hw.buf)
		hash := hex.EncodeToString(mac.Sum(nil))
		hw.ResponseWriter.Header().Set("HashSHA256", hash)
	}

	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	hw.ResponseWriter.WriteHeader(hw.status)
	if len(hw.buf) > 0 {
		hw.ResponseWriter.Write(hw.buf)
	}
}

func RequestSignatureCheck(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inputHashStr := r.Header.Get("HashSHA256")
			if inputHashStr == "" {
				next.ServeHTTP(w, r)
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

func ResponseSignatureAdd(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				next.ServeHTTP(w, r)
				return
			}

			hw := &hashWriter{
				ResponseWriter: w,
				secret:         secret,
			}

			next.ServeHTTP(hw, r)
			hw.writeHashedResponse()
		})
	}
}
