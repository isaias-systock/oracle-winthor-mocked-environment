package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginReturnsDifferentTokens(t *testing.T) {
	var tokens []string
	for range 2 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, loginPath, nil)

		loginHandler(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var response struct {
			Token string `json:"accessToken"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Token) != 64 {
			t.Fatalf("token length = %d, want 64", len(response.Token))
		}
		tokens = append(tokens, response.Token)
	}

	if tokens[0] == tokens[1] {
		t.Fatal("two login requests returned the same token")
	}
}

func TestLoginRejectsNonPost(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, loginPath, nil)

	loginHandler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func TestSaleEchoesValidJSON(t *testing.T) {
	body := `{"numPedRca":123,"items":[]}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		salePath+"?ignoraProcessamento=false",
		strings.NewReader(body),
	)

	saleHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response saleResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Message != "Pedidos gerados com sucesso!" {
		t.Fatalf("message = %q", response.Message)
	}
	if string(response.Data) != body {
		t.Fatalf("data = %s, want %s", response.Data, body)
	}
}

func TestSaleRejectsInvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, salePath, strings.NewReader(`{"`))

	saleHandler(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
