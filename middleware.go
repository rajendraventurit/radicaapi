package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rajendraventurit/radicaapi/lib/logger"
	"github.com/rajendraventurit/radicaapi/lib/token"
)

var localDB *sqlx.DB

const methodOpt = "OPTIONS"

// SetLocalDB sets the DB for use in the middleware
func SetLocalDB(db *sqlx.DB) {
	localDB = db
}

func newLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.HTTPAccess(0, r)
		next.ServeHTTP(w, r)
	})
}

func newHeaderHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%v", 3600*time.Second))
		w.Header().Set("Access-Control-Allow-Methods", rTable.Methods(r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

func newTokenHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToUpper(r.Method) == methodOpt {
			next.ServeHTTP(w, r)
			return
		}
		route, err := rTable.GetRoute(r.Method, r.URL.Path)
		if err != nil {
			logger.Errorf("Route not found %v", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if route.Insecure {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := token.AuthToken(r); err != nil {
			logger.Errorf(err.Error())
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newPermMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToUpper(r.Method) == methodOpt {
			next.ServeHTTP(w, r)
			return
		}
		route, err := rTable.GetRoute(r.Method, r.URL.Path)
		if err != nil {
			logger.Errorf("Route not found %v", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if route.Insecure {
			next.ServeHTTP(w, r)
			return
		}
		claim, err := token.AuthToken(r)
		if err != nil {
			logger.Errorf(err.Error())
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		logger.Errorf("Unauthorized UserID %v", claim.UserID)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func getOrgID(r *http.Request) (int64, error) {
	if r.Method == "GET" {
		o := r.URL.Query().Get("organization_id")
		return strconv.ParseInt(o, 10, 64)
	}
	var bodyBytes []byte
	if r.Body == nil {
		return 0, fmt.Errorf("Body is nil")
	}
	bodyBytes, _ = ioutil.ReadAll(r.Body)
	// Restore the io.ReadCloser to its original state
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	mp := make(map[string]interface{})
	if err := json.Unmarshal(bodyBytes, &mp); err != nil {
		return 0, fmt.Errorf("Failed to decode json")
	}
	o, ok := mp["organization_id"]
	if !ok {
		return 0, fmt.Errorf("organization_id not found in JSON")
	}
	return strconv.ParseInt(fmt.Sprintf("%v", o), 10, 64)
}
