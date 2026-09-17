package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestZibalGatewayRequestAndVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if body["merchant"] != "merchant-code" {
			t.Errorf("merchant=%v", body["merchant"])
		}
		switch r.URL.Path {
		case "/v1/request":
			if body["amount"] != float64(123000) {
				t.Errorf("amount=%v", body["amount"])
			}
			if body["orderId"] != "order-1" {
				t.Errorf("orderId=%v", body["orderId"])
			}
			_, _ = w.Write([]byte(`{"result":100,"trackId":123456}`))
		case "/v1/verify":
			if body["trackId"] != float64(123456) {
				t.Errorf("trackId=%v", body["trackId"])
			}
			_, _ = w.Write([]byte(`{"result":201,"amount":123000,"refNumber":987654321}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	gateway := NewZibalGateway("merchant-code", server.URL)
	track, err := gateway.Request(context.Background(), 123000, "https://clubmax.ir/callback", "order-1", "tuition", "09120000000")
	if err != nil || track != 123456 {
		t.Fatalf("request track=%d err=%v", track, err)
	}
	result, err := gateway.Verify(context.Background(), track)
	if err != nil {
		t.Fatal(err)
	}
	if result.Amount != 123000 || result.RefNumber != "987654321" {
		t.Fatalf("verify=%+v", result)
	}
	if got := gateway.StartURL(track); got != server.URL+"/start/123456" {
		t.Fatalf("start URL=%q", got)
	}
}

func TestZibalGatewayRejectsRedirectsAndIncompleteVerification(t *testing.T) {
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("payment client followed an HTTP redirect")
	}))
	defer redirectTarget.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL, http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	gateway := NewZibalGateway("merchant-code", redirector.URL)
	if _, err := gateway.Request(context.Background(), 123000, "https://clubmax.ir/callback", "order-1", "tuition", ""); err == nil {
		t.Fatal("expected redirects to be rejected")
	}

	incomplete := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":100,"amount":123000}`))
	}))
	defer incomplete.Close()
	gateway = NewZibalGateway("merchant-code", incomplete.URL)
	if _, err := gateway.Verify(context.Background(), 123456); err == nil {
		t.Fatal("expected incomplete verification to be rejected")
	}
}
