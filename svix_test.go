package svix_plugin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	svix "github.com/svix/svix-webhooks/go"
)

func TestSvix(t *testing.T) {
	// testing variables (from: https://github.com/svix/svix-webhooks/blob/efa8ef1d179acba49afac1e8c156527856fdf787/go/webhook_test.go#L212)
	secret := "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"
	msgID := "msg_p5jXN8AQM9LWM0D4loKWxJek"
	timestamp := time.Now()
	payload := []byte(`{"test": 2432232314}`)

	wh, err := svix.NewWebhook(secret)

	if err != nil {
		t.Fatalf("Error creating svix webhook: %s", err)
	}

	signature, err := wh.Sign(msgID, timestamp, payload)
	if err != nil {
		t.Fatalf("Error creating svix signature: %s", err)
	}

	cfg := CreateConfig()
	cfg.secret = secret

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "svix-plugin")
	if err != nil {
		t.Fatalf("Error intializing svix plugin: %s", err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", bytes.NewBuffer(payload))

	if err != nil {
		t.Fatalf("Error during request: %s", err)
	}

	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", strconv.FormatInt(timestamp.Unix(), 10))
	req.Header.Set("svix-signature", signature)

	handler.ServeHTTP(recorder, req)

	assertAllowed(t, recorder.Result())
}

func assertAllowed(t *testing.T, res *http.Response) {
	t.Helper()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		t.Errorf("Invalid response code: %d", res.StatusCode)
	}
}
