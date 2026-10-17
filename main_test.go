package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendForwardsCaptchaWidgetRecordID(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/captcha/verify" {
			t.Errorf("unexpected API path: %s", r.URL.Path)
			return
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if body["widget_record_id"] != "widget-test" || body["token"] != "invalid-test-token" || body["action"] != "login" {
			t.Errorf("unexpected captcha request: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"ok":false,"error":{"code":"INVALID_TOKEN","message":"test token"}}`))
	}))
	defer api.Close()
	s := &server{client: &InfraiClient{BaseURL: api.URL, Key: "test", HTTP: api.Client()}, sellers: map[string]string{}}
	req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader(`{"phone":"not-a-phone","token":"invalid-test-token","widget_record_id":"widget-test"}`))
	rec := httptest.NewRecorder()
	s.send(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestVerifyCodePublishesHandoff(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer api.Close()
	s := &server{client: &InfraiClient{BaseURL: api.URL, Key: "test", HTTP: api.Client()}, sellers: map[string]string{}}
	req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(`{"phone":"+15551234567","code":"123456"}`))
	rec := httptest.NewRecorder()
	s.verify(rec, req)
	if rec.Code != 200 || s.sellers["+15551234567"] != "verified" {
		t.Fatalf("handoff not published: code=%d sellers=%v", rec.Code, s.sellers)
	}
}
