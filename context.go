package zeal

import (
	"encoding/json"
	"net/http"
)

type Writer[ResponseType any] struct {
	http.ResponseWriter
}

func (w *Writer[ResponseType]) JSON(data ResponseType, statuses ...int) {
	w.Header().Add("Content-Type", "application/json")
	if len(statuses) > 0 {
		status := statuses[0]
		w.WriteHeader(status)
	}
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
