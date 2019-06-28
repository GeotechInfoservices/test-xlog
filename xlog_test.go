package xlog

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xsided/h8tp/request"
	"github.com/xsided/h8tp/response"
)

func TestRequestLogger(t *testing.T) {
	handler := WithRequestLogger(func(ctx context.Context, req request.Request) (events.APIGatewayProxyResponse, error) {
		logger, err := GetLogger(ctx)
		logger.Logger.SetFormatter(&logrus.TextFormatter{})
		memLog := &bytes.Buffer{}
		logger.Logger.SetOutput(memLog)
		if err != nil {
			fmt.Println("Error while trying to get logger", err)
			t.Log()
			t.Fail()
		}

		logger.Info("Test", map[string]interface{}{
			"some": "message",
			"else": ctx.Value(Logger),
		})

		if !strings.Contains(string(memLog.Bytes()), "trace_id=some-trace-id") {
			fmt.Println("Log doesn't contain trace information", err)
			t.Log()
			t.Fail()
		}

		if !strings.Contains(string(memLog.Bytes()), "user_id=some-user-id") {
			fmt.Println("Log doesn't contain user information", err)
			t.Log()
			t.Fail()
		}

		return response.OK("test")
	})

	handler(context.Background(), request.Request{
		Headers: map[string]string{
			"x-trace-id": "some-trace-id",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"subject":  "some-user-id",
				"owner_id": "some-owner-1",
			},
		},
	})
}
