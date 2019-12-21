package handler

import (
	"net/http"

	"github.com/rajendraventurit/radicaapi/lib/env"
	"github.com/rajendraventurit/radicaapi/lib/serror"
)

// HFunc is a handler function
type HFunc func(e *env.Env, w http.ResponseWriter, r *http.Request) error

// Handler for http handlers
type Handler struct {
	Env *env.Env
	Fn  HFunc
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Fn(h.Env, w, r)
	if err != nil {
		switch err.(type) {
		case serror.Errorer:
			err.(serror.Error).Log(0, r)
			err.(serror.Error).Send(w)
		default:
			se := serror.NewCode(http.StatusInternalServerError, err)
			se.Log(0, r)
			se.Send(w)
		}
	}
}
