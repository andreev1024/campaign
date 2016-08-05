package res

import (
	"net/http"
)

type Message struct {
	Message string `json:"message" required:"true"`
}

// Error struct represents error response
type Error struct {
	Error interface{} `json:"error"`
}

// CreateMessageError function creates message error response
func CreateMessageError(message string) Error {
	return Error{Message{message}}
}

// Forbidden function returns forbidden error (403)
func Forbidden(messages ...string) Error {
	return createMessageError(http.StatusText(http.StatusForbidden), messages)
}

func NotFound(messages ...string) Error {
	return createMessageError(http.StatusText(http.StatusNotFound), messages)
}

func InternalServerError(messages ...string) Error {
	return createMessageError(http.StatusText(http.StatusInternalServerError), messages)
}

func BadRequest(messages ...string) Error {
	return createMessageError(http.StatusText(http.StatusBadRequest), messages)
}

func MethodNotAllowed(messages ...string) Error {
	return createMessageError(http.StatusText(http.StatusMethodNotAllowed), messages)
}

func createMessageError(defaultMessage string, messages []string) Error {
	if len(messages) > 0 {
		return CreateMessageError(messages[0])
	}
	return CreateMessageError(defaultMessage)
}

func Ok(messages ...string) Message {
	return createMessage(http.StatusText(http.StatusOK), messages)
}

func createMessage(defaultMessage string, messages []string) Message {
	if len(messages) > 0 {
		return Message{messages[0]}
	}
	return Message{defaultMessage}
}
