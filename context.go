package zeal

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Response[T_Response any] struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (r *Response[T_Response]) JSON(data T_Response, status ...int) {
	r.ResponseWriter.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(r.ResponseWriter).Encode(data); err != nil {
		r.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(status) > 0 {
		r.ResponseWriter.WriteHeader(status[0])
	}
}

func (r Response[T_Response]) Status(status int) {
	r.ResponseWriter.WriteHeader(status)
}

func (r Response[T_Response]) Error(status int, errorMsg ...string) {
	if len(errorMsg) > 0 {
		http.Error(r.ResponseWriter, errorMsg[0], ensureErrorCode(status))
		return
	}

	http.Error(r.ResponseWriter, http.StatusText(status), ensureErrorCode(status))
}

func ensureErrorCode(status int) int {
	codeStr := strconv.Itoa(status)
	if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
		return status
	}

	return 500
}
