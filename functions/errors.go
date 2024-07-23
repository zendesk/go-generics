package functions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

func HasNonNilError(errs []error) bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

func MergeErrors(errs ...error) error {
	if len(errs) == 0 || !HasNonNilError(errs) {
		return nil
	}

	var messages []string

	for i := len(errs) - 1; i >= 0; i-- {
		if errs[i] != nil {
			messages = append(messages, errs[i].Error())
		}
	}

	return errors.New(strings.Join(messages, ", "))
}

// WrapError prepends a message to an error, and returns a new error while preserving its type.
func WrapError(err error, message string) error {
	if err == nil {
		return fmt.Errorf(message)
	}
	return fmt.Errorf("%s; %w", message, err)
}

// WrapErrorf performs the same action as WrapError, but allows the user to format the message (similar to fmt.Printf).
func WrapErrorf(err error, message string, args ...any) error {
	resolvedMessage := fmt.Sprintf(message, args...)
	return WrapError(err, resolvedMessage)
}

// WrapErrorsIntoFirst wraps all errors into the first preserving type.
func WrapErrorsIntoFirst(errs ...error) error {
	if len(errs) == 0 || !utils.HasNonNilError(errs) {
		return nil
	}

	err := errs[0]

	for i := 1; i < len(errs); i++ {
		err = WrapError(err, errs[i].Error())
	}
	return err
}
