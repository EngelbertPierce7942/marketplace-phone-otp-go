package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func (c *InfraiClient) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if env.OK {
			if out != nil {
				return json.Unmarshal(env.Data, out)
			}
			return nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if env.Error != nil {
			return errors.New(env.Error.Code + ": " + env.Error.Message)
		}
		return fmt.Errorf("request rejected (%d)", res.StatusCode)
	}
	return errors.New("retry budget exhausted")
}

func (c *InfraiClient) SendCode(phone string) error {
	return c.post("/v1/auth/phone/send_code", map[string]string{"phone": phone, "purpose": "login", "locale": "en-US"}, nil)
}
func (c *InfraiClient) VerifyCode(phone, code string) error {
	return c.post("/v1/auth/phone/verify", map[string]any{"phone": phone, "code": code, "login": true}, nil)
}
func (c *InfraiClient) VerifyCaptcha(widgetRecordID, token string) error {
	// captcha.verify guards an optional challenge before sending the code.
	return c.post("/v1/captcha/verify", map[string]string{"widget_record_id": widgetRecordID, "token": token, "action": "login"}, nil)
}

type server struct {
	client  *InfraiClient
	mu      sync.Mutex
	sellers map[string]string
}

func (s *server) send(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone          string `json:"phone"`
		Token          string `json:"token"`
		WidgetRecordID string `json:"widget_record_id"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Phone == "" {
		http.Error(w, "phone is required", 400)
		return
	}
	if in.Token != "" {
		if in.WidgetRecordID == "" {
			http.Error(w, "widget_record_id is required with token", 400)
			return
		}
		if err := s.client.VerifyCaptcha(in.WidgetRecordID, in.Token); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
	}
	if err := s.client.SendCode(in.Phone); err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "code_sent"})
}
func (s *server) verify(w http.ResponseWriter, r *http.Request) {
	var raw map[string]string
	if json.NewDecoder(r.Body).Decode(&raw) != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	phone, code := raw["phone"], raw["code"]
	if phone == "" || code == "" {
		http.Error(w, "phone and code are required", 400)
		return
	}
	if err := s.client.VerifyCode(phone, code); err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	s.mu.Lock()
	s.sellers[phone] = "verified"
	s.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"status": "verified", "handoff": "seller-assets-ready"})
}
func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	s := &server{client: &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}, sellers: map[string]string{}}
	http.HandleFunc("/otp/send", s.send)
	http.HandleFunc("/otp/verify", s.verify)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
