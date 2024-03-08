package zeal

import (
	"encoding/json"
	"net/http"
)

type ResponseWriter[ResponseType any] struct {
	http.ResponseWriter
}

func (w *ResponseWriter[ResponseType]) JSON(data ResponseType, status ...int) {
	w.Header().Add("Content-Type", "application/json")
	if len(status) > 0 {
		firstStatus := status[0]
		w.WriteHeader(firstStatus)
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type Request[ParamsType any] struct {
	*http.Request
	Params ParamsType
}
