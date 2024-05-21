package ldevents

import (
	"fmt"
)

func httpErrorMessage(statusCode int, context string, recoverableMessage string) string {
	statusDesc := ""
	if statusCode == 401 {
		statusDesc = " (invalid SDK key)"
	}
	resultMessage := recoverableMessage
	if !isHTTPErrorRecoverable(statusCode) {
		resultMessage = "giving up permanently"
	}
	return fmt.Sprintf("Received HTTP error %d%s for %s - %s",
		statusCode, statusDesc, context, resultMessage)
}

// Tests whether an HTTP error status represents a condition that might resolve on its own if we retry,
// or at least should not make us permanently stop sending requests.
func isHTTPErrorRecoverable(statusCode int) bool {
	if statusCode >= 400 && statusCode < 500 {
		switch statusCode {
		case 400: // bad request
			return true
		case 408: // request timeout
			return true
		case 429: // too many requests
			return true
		default:
			return false // all other 4xx errors are unrecoverable
		}
	}
	return true
}
