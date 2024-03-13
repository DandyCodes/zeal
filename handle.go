package zeal

import (
	"net/http"
)

type HandlerFunc[T_Response, T_Params, T_Body any] func(Response[T_Response], T_Params, T_Body)

func Handle[T_Response, T_Params, T_Body any](mux *ServeMux, pattern string, handlerFunc HandlerFunc[T_Response, T_Params, T_Body]) {
	registerRoute[T_Response, T_Params, T_Body](pattern, mux)
	mux.ServeMux.HandleFunc(pattern, unwrapHandlerFunc(handlerFunc))
}

func unwrapHandlerFunc[T_Response, T_Params, T_Body any](handlerFunc HandlerFunc[T_Response, T_Params, T_Body]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, body, err := getArgs[T_Params, T_Body](r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}

		handlerFunc(Response[T_Response]{w, r}, params, body)
	}
}
