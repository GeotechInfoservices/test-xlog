package xlog

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"dev.azure.com/ManyDigital/MDLiveQuiz/_git/livequiz-h8tp/request"
	"dev.azure.com/ManyDigital/MDLiveQuiz/_git/livequiz-h8tp/response"
	"github.com/aws/aws-lambda-go/events"
)

func TestRequestLogger(t *testing.T) {
	handler := WithRequestLogger(func(ctx context.Context, req request.Request) (response.Response, error) {
		logger := GetLogger(ctx)
		logger.Logger.SetFormatter(&logrus.TextFormatter{})
		memLog := &bytes.Buffer{}
		logger.Logger.SetOutput(memLog)

		logger.Info("Test", map[string]interface{}{
			"some": "message",
			"else": ctx.Value(Logger),
		})

		if !strings.Contains(string(memLog.Bytes()), "trace_id=some-trace-id") {
			fmt.Println("Log doesn't contain trace information")
			t.Log()
			t.Fail()
		}

		if !strings.Contains(string(memLog.Bytes()), "user_id=some-user-id") {
			fmt.Println("Log doesn't contain user information")
			t.Log()
			t.Fail()
		}

		return response.OK("test")
	})

	resp, err := handler(context.Background(), request.Request{
		Headers: map[string]string{
			"x-trace-id": "some-trace-id",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"principalId": "some-user-id",
				"owner_id":    "some-owner-1",
			},
		},
	})

	if err != nil {
		fmt.Printf("Test failed with: %s", err)
		t.Log()
		t.Fail()
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Did not end with 200")
		t.Log()
		t.Fail()
	}
}
