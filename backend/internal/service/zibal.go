package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
)

// PaymentGateway keeps payment-provider details out of the finance workflow.
// Amounts are always in rials.
type PaymentGateway interface {
	Request(context.Context, int64, string, string, string, string) (int64, error)
	Verify(context.Context, int64) (GatewayVerification, error)
	StartURL(int64) string
}

type GatewayVerification struct {
	Amount    int64
	RefNumber string
}

type ZibalGateway struct {
	merchant string
	baseURL  string
	client   *http.Client
}

func NewZibalGateway(merchant, baseURL string) *ZibalGateway {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://gateway.zibal.ir"
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return &ZibalGateway{
		merchant: strings.TrimSpace(merchant),
		baseURL:  strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout:       15 * time.Second,
			Transport:     transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

type zibalResponse struct {
	Result    int64          `json:"result"`
	Message   string         `json:"message"`
	TrackID   int64          `json:"trackId"`
	Amount    int64          `json:"amount"`
	RefNumber zibalReference `json:"refNumber"`
}

// Zibal integrations in the wild return refNumber as either a JSON string or
// number. Preserve its exact decimal representation instead of converting it
// through float64, which could silently lose digits.
type zibalReference string

func (v *zibalReference) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*v = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = zibalReference(text)
		return nil
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil {
		return fmt.Errorf("invalid refNumber: %w", err)
	}
	*v = zibalReference(number.String())
	return nil
}

func (g *ZibalGateway) Request(ctx context.Context, amount int64, callbackURL, orderID, description, mobile string) (int64, error) {
	if g.merchant == "" || amount <= 0 || callbackURL == "" {
		return 0, domain.ErrPaymentGateway
	}
	body := map[string]any{"merchant": g.merchant, "amount": amount, "callbackUrl": callbackURL, "orderId": orderID, "description": description}
	if mobile != "" {
		body["mobile"] = mobile
	}
	var out zibalResponse
	if err := g.post(ctx, "/v1/request", body, &out); err != nil {
		return 0, err
	}
	if out.Result != 100 || out.TrackID <= 0 {
		return 0, fmt.Errorf("%w: zibal request result %d: %s", domain.ErrPaymentGateway, out.Result, out.Message)
	}
	return out.TrackID, nil
}

func (g *ZibalGateway) Verify(ctx context.Context, trackID int64) (GatewayVerification, error) {
	if g.merchant == "" || trackID <= 0 {
		return GatewayVerification{}, domain.ErrPaymentGateway
	}
	var out zibalResponse
	if err := g.post(ctx, "/v1/verify", map[string]any{"merchant": g.merchant, "trackId": trackID}, &out); err != nil {
		return GatewayVerification{}, err
	}
	if out.Result != 100 && out.Result != 201 {
		return GatewayVerification{}, fmt.Errorf("%w: zibal verify result %d: %s", domain.ErrPaymentFailed, out.Result, out.Message)
	}
	refNumber := strings.TrimSpace(string(out.RefNumber))
	if out.Amount <= 0 || refNumber == "" || len(refNumber) > 100 {
		return GatewayVerification{}, fmt.Errorf("%w: incomplete zibal verification response", domain.ErrPaymentGateway)
	}
	return GatewayVerification{Amount: out.Amount, RefNumber: refNumber}, nil
}

func (g *ZibalGateway) post(ctx context.Context, path string, input, output any) error {
	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrPaymentGateway, err)
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(output); err != nil {
		return fmt.Errorf("%w: invalid zibal response: %v", domain.ErrPaymentGateway, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: zibal returned HTTP %d", domain.ErrPaymentGateway, resp.StatusCode)
	}
	return nil
}

func (g *ZibalGateway) StartURL(trackID int64) string {
	return fmt.Sprintf("%s/start/%d", g.baseURL, trackID)
}
