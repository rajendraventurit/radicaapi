package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rajendraventurit/radicaapi/lib/serror"
)

func decodeJSON(r io.Reader, i interface{}) error {
	if err := json.NewDecoder(r).Decode(i); err != nil {
		return serror.New(http.StatusBadRequest, err, "json.Decode", err.Error())
	}
	return nil
}

func sendJSON(w io.Writer, i interface{}) error {
	js, err := json.Marshal(&i)
	if err != nil {
		return serror.New(http.StatusInternalServerError, err, "json.Marshal")
	}

	_, err = w.Write(js)
	return err
}

func sendJSON1(w http.ResponseWriter, i interface{}, success bool, message string, code int) error {
	//js, err := json.Marshal(&i)

	resp := serror.NewResponseJSON(success, message, i, code)

	respobj, err := json.Marshal(&resp)
	if err != nil {
		return serror.New(http.StatusInternalServerError, err, "json.Marshal")
	}

	w.WriteHeader(code)
	_, err = w.Write(respobj)
	return err
}

func getQueryInt64(r *http.Request, key string) (int64, error) {
	val := r.URL.Query().Get(key)
	in, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, serror.New(http.StatusBadRequest, err, "ParseInt", fmt.Sprintf("%s is an invalid %s", val, key))
	}
	return in, nil
}

func getQueryBool(r *http.Request, key string) (bool, error) {
	val := r.URL.Query().Get(key)
	bl, err := strconv.ParseBool(val)
	if err != nil {
		return false, serror.New(http.StatusBadRequest, err, "ParseBool", fmt.Sprintf("%s is an invalid %s", val, key))
	}
	return bl, nil
}
