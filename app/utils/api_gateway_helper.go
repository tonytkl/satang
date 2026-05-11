package utils

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

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
