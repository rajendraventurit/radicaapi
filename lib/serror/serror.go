package serror

import (
	"fmt"
	"net/http"

	"github.com/rajendraventurit/radicaapi/lib/logger"
)

// Errorer is a handler error
type Errorer interface {
	error
	Status() string
	StatusCode() int
	Message() string
}

// Error is an error associated with an http status
type Error struct {
	Code    int
	Err     error
	Context string
	Msg     string // User message sent with html error
}

//ResponseJSON is the struct
type ResponseJSON struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Code    int         `json:"code"`
}

//NewResponseJSON which returns the response
func NewResponseJSON(status bool, message string, data interface{}, code int) ResponseJSON {
	return ResponseJSON{Status: status, Message: message, Data: data, Code: code}
}

// New returns a new Error
// code = http status code
// err = error
// con = context
// msg = user message
func New(code int, err error, con string, userMsg ...string) Error {
	if len(userMsg) > 0 && userMsg[0] != "" {

		return Error{Code: code, Err: err, Context: con, Msg: userMsg[0]}
	}

	return Error{Code: code, Err: err, Context: con}
}

// NewCode returns a new Error with a http status code and error
func NewCode(code int, err error) Error {
	return Error{Code: code, Err: err}
}

// NewServer will return a new internal server error
func NewServer(err error, con string, userMsg ...string) Error {
	return New(http.StatusInternalServerError, err, con, userMsg...)
}

// NewBadRequest will return a new internal server error
func NewBadRequest(err error, con string, userMsg ...string) Error {
	return New(http.StatusBadRequest, err, con, userMsg...)
}

// Error satisfies the error interface
func (e Error) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%v %v", e.Context, e.Err)
	}
	return e.Err.Error()
}

// StatusCode returns the http status code
func (e Error) StatusCode() int {
	return e.Code
}

// Status returns the string associated with a status code
func (e Error) Status() string {
	return http.StatusText(e.Code)
}

// Message returns a user message
func (e Error) Message() string {
	return e.Msg
}

// Send will send the status code and description to the response writer
func (e Error) Send(w http.ResponseWriter) {
	msg := fmt.Sprintf("%s %s", e.Status(), e.Message())
	http.Error(w, msg, e.StatusCode())
}

// Log write to the logger
func (e Error) Log(userid int64, r *http.Request) {
	msg := ""
	switch {
	case r == nil && userid == 0:
		msg = e.Error()
	case r == nil:
		msg = fmt.Sprintf("UserID %v %v", userid, e.Error())
	case userid == 0:
		msg = fmt.Sprintf("%s %v %v", r.Method, r.URL.String(), e.Error())
	default:
		msg = fmt.Sprintf("UserID %v %v %v %v", userid, r.Method, r.URL.String(), e.Error())
	}
	logger.Error(msg)
}
