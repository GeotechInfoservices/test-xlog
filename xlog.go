package xlog

import (
	"context"
	"fmt"
	"net/http"

	"dev.azure.com/ManyDigital/MDLiveQuiz/_git/livequiz-h8tp/request"
	"dev.azure.com/ManyDigital/MDLiveQuiz/_git/livequiz-h8tp/response"
	"github.com/sirupsen/logrus"
)

// NewRequestLogger Returns an instance of XLog which is set up to log within an http request
// This requires a trace id and the user id.
func NewRequestLogger(trace string, userID string, ownerID string) *XLog {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &XLog{
		TracingID: trace,
		UserID:    userID,
		OwnerID:   ownerID,
		Logger:    logger,
	}
}

// NewLogger Returns a clean instance of XLog
func NewLogger() *XLog {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &XLog{
		Logger: logger,
	}
}

// XLog is a structuted logger with an opiniated format.
// Must use this log format to be consistent with the underlying infrastructure.
type XLog struct {
	TracingID string
	UserID    string
	OwnerID   string
	Logger    *logrus.Logger
}

// Fields for a logger
type Fields map[string]interface{}

// Error logs an error in xsided format
func (x *XLog) Error(err error, msg string, args Fields) {

	fields := logrus.Fields{
		"error":    err,
		"trace_id": x.TracingID,
		"user_id":  x.UserID,
		"owner_id": x.OwnerID,
	}

	for k, v := range args {
		fields[k] = v
	}

	x.Logger.WithFields(fields).Error(msg)
}

// Info logs info level log in xsided format
func (x *XLog) Info(msg string, args Fields) {

	fields := logrus.Fields{
		"trace_id": x.TracingID,
		"user_id":  x.UserID,
		"owner_id": x.OwnerID,
	}

	for k, v := range args {
		fields[k] = v
	}

	x.Logger.WithFields(fields).Info(msg)
}

type key string

// Logger key used to fetch the logger from the context
const (
	Logger key = "logger"
)

// WithRequestLogger provides a request logger to the requests context
// Pulls required information from the request context and instantiates a new logger
func WithRequestLogger(h func(context.Context, request.Request) (response.Response, error)) func(context.Context, request.Request) (response.Response, error) {
	return func(ctx context.Context, req request.Request) (response.Response, error) {

		ownerid, ok := req.RequestContext.Authorizer["owner_id"]
		if !ok {
			return response.InvalidRequest("misconfigured logger")
		}

		userid, ok := req.RequestContext.Authorizer["principalId"]
		if !ok {
			return response.InvalidRequest("misconfigured logger")
		}

		header := http.Header{}
		for key, val := range req.Headers {
			header.Add(key, val)
		}

		trace := header.Get("X-Trace-Id")
		if trace == "" {
			return response.InvalidRequest("Must provide X-Trace-Id header in request")
		}

		l := NewRequestLogger(trace, userid.(string), ownerid.(string))
		c := context.WithValue(ctx, Logger, l)
		return h(c, req)
	}
}

// GetLogger returns a logger from a context
func GetLogger(ctx context.Context) *XLog {
	logger := ctx.Value(Logger)

	switch logger.(type) {
	case *XLog:
		return logger.(*XLog)
	default:
		panic(fmt.Sprintf("Misconfigured logger: %+v", logger))
	}
}
