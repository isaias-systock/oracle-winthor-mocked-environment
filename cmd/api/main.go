package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

const (
	loginPath = "/winthor/autenticacao/v1/login"
	salePath  = "/winthor/venda/v0/importar-venda"
	listenOn  = ":1522"
)

type saleResponse struct {
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(loginPath, loginHandler)
	mux.HandleFunc(salePath, saleHandler)

	log.Printf("API mock Winthor ouvindo em %s", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, mux))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "não foi possível gerar o token",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"accessToken": hex.EncodeToString(token),
	})
}

func saleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
	if err != nil || len(body) > 1<<20 || !json.Valid(body) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"message": "o corpo deve conter um JSON válido",
		})
		return
	}

	writeJSON(w, http.StatusOK, saleResponse{
		Message: "Pedidos gerados com sucesso!",
		Data:    body,
	})
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", http.MethodPost)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
		"message": "método não permitido",
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("falha ao escrever resposta JSON: %v", err)
	}
}
