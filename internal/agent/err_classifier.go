package agent

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type agentErrorClassifier struct{}

func NewAgentErrorClassifier() *agentErrorClassifier {
	return &agentErrorClassifier{}
}

type SendError struct {
	Msg  string
	Code int
	err  error
}

func NewSendError(code int, format string, a ...any) error {
	return &SendError{
		Msg:  fmt.Sprintf(format, a...),
		Code: code,
	}
}

func WrapSendError(code int, msg string, err error) error {
	return &SendError{
		Msg:  msg,
		Code: code,
		err:  err,
	}
}

func (e *SendError) Error() string {
	if e.err != nil && e.err.Error() != e.Error() {
		return fmt.Sprintf("%s (code: %d, previous: %v)", e.Msg, e.Code, e.err)
	}
	return fmt.Sprintf("%s (code: %d)", e.Msg, e.Code)
}

func (e *SendError) Unwrap() error {
	return e.err
}

func (c *agentErrorClassifier) IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "network is unreachable") ||
		strings.Contains(err.Error(), "EOF") {
		return true
	}

	var httpErr *SendError
	if errors.As(err, &httpErr) {
		code := httpErr.Code
		if code >= 500 && code < 600 {
			return true
		}
		if code == http.StatusRequestTimeout || code == http.StatusTooManyRequests { // Too Many Requests
			return true
		}
	}

	return false
}
