package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

type validationError struct {
	err error
}

var ErrInvalidCursor = errors.New("Invalid cursor")

func (e validationError) Error() string { return e.err.Error() }
func (e validationError) Unwrap() error { return e.err }

func ValidationError(err error) error {
	return validationError{err: err}
}

func InvalidCursorError(err error) error {
	return ValidationError(fmt.Errorf("%w: %w", ErrInvalidCursor, err))
}

func ErrorMessage(err error) string {
	if errors.Is(err, ErrInvalidCursor) {
		return ErrInvalidCursor.Error()
	}
	return err.Error()
}

// ErrorStatusCode maps classified input and upstream errors to API Gateway statuses.
func ErrorStatusCode(err error) int {
	var validation validationError
	if errors.As(err, &validation) {
		return http.StatusBadRequest
	}

	var apiError interface{ ErrorCode() string }
	if errors.As(err, &apiError) {
		if _, ok := retry.DefaultThrottleErrorCodes[apiError.ErrorCode()]; ok {
			return http.StatusTooManyRequests
		}
		if _, ok := retry.DefaultRetryableErrorCodes[apiError.ErrorCode()]; ok {
			return http.StatusRequestTimeout
		}
	}

	var responseError interface{ HTTPStatusCode() int }
	if errors.As(err, &responseError) {
		status := responseError.HTTPStatusCode()
		if status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || (status >= 500 && status <= 599) {
			return status
		}
	}

	var timeout interface{ Timeout() bool }
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

func JsonResponse(statusCode int, payload any) (events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, fmt.Errorf("marshal response: %w", err)
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}
