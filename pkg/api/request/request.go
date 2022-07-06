package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Do runs a get request against the given url.
func Do(ctx context.Context, url string) (int, string, error) {
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   1 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, string(body), nil
	}

	return 0, "", fmt.Errorf("expected a status code between 200 and 300, but got: %d", resp.StatusCode)
}
