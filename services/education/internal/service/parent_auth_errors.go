package service

import "strings"

var ErrParentReauthRequired = &parentReauthRequiredError{}

type parentReauthRequiredError struct {
	message string
}

func (e *parentReauthRequiredError) Error() string {
	message := strings.TrimSpace(e.message)
	if message == "" {
		return "当前登录态已失效，请重新登录"
	}
	return message
}

func (e *parentReauthRequiredError) Is(target error) bool {
	_, ok := target.(*parentReauthRequiredError)
	return ok
}

func newParentReauthRequiredError(message string) error {
	return &parentReauthRequiredError{message: message}
}
