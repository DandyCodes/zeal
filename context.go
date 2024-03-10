package zeal

import (
	"encoding/json"
	"net/http"
)

type ResponseWriter[T_Response any] struct {
	http.ResponseWriter
}

func (w *ResponseWriter[T_Response]) JSON(data T_Response, status ...int) {
	w.Header().Add("Content-Type", "application/json")
	if len(status) > 0 {
		firstStatus := status[0]
		w.WriteHeader(firstStatus)
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type Request[T_Params any] struct {
	*http.Request
	Params T_Params
}
