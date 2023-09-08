package zeal

import (
	"encoding/json"
	"net/http"
)

type Writer[ResponseType any] struct {
	http.ResponseWriter
}

func (w *Writer[ResponseType]) JSON(status int, data ResponseType) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type Rqr[ParamsType any] struct {
	Request *http.Request
	Params  ParamsType
}

type Rqw[ParamsType, BodyType any] struct {
	Rqr[ParamsType]
	Body BodyType
}
