package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/bunrouter"
)

func UnmarshalJSON(
	w http.ResponseWriter,
	req bunrouter.Request,
	dst interface{},
	maxBytes int64,
) error {
	req.Body = http.MaxBytesReader(w, req.Body, maxBytes)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func JSON(w http.ResponseWriter, value interface{}, s int) error {
	if s == 0 {
		s = http.StatusOK
	}
	w.WriteHeader(s)

	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}

func BindJSON(
	w http.ResponseWriter,
	req bunrouter.Request,
	dst interface{},
) error {
	if err := UnmarshalJSON(w, req, dst, 10); err != nil {
		return err
	}
	return nil
}
