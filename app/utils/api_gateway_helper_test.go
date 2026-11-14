package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestErrorStatusCode(t *testing.T) {
	upstream := func(status int, cause error) error {
		return fmt.Errorf("repository: %w", &smithyhttp.ResponseError{
			Response: &smithyhttp.Response{Response: &http.Response{StatusCode: status}},
			Err:      cause,
		})
	}
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"validation", fmt.Errorf("service: %w", ValidationError(errors.New("invalid input"))), 400},
		{"bad upstream request is internal", upstream(400, errors.New("bad request")), 500},
		{"upstream not found is internal", upstream(404, errors.New("not found")), 500},
		{"request timeout", upstream(408, errors.New("timeout")), 408},
		{"too many requests", upstream(429, errors.New("throttled")), 429},
		{"internal server error", upstream(500, errors.New("failed")), 500},
		{"bad gateway", upstream(502, errors.New("failed")), 502},
		{"service unavailable", upstream(503, errors.New("failed")), 503},
		{"gateway timeout", upstream(504, errors.New("failed")), 504},
		{"other upstream server error", upstream(507, errors.New("failed")), 507},
		{"dynamodb throttling", upstream(400, &smithy.GenericAPIError{Code: "ProvisionedThroughputExceededException"}), 429},
		{"aws timeout", upstream(400, &smithy.GenericAPIError{Code: "RequestTimeoutException"}), 408},
		{"context deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), 504},
		{"ordinary error", errors.New("database failed"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ErrorStatusCode(tt.err))
		})
	}
}
