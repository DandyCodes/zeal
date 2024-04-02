package zeal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Response[T_Response any] struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (r Response[T_Response]) Status(status int) error {
	r.ResponseWriter.WriteHeader(status)
	return nil
}

func (r Response[T_Response]) Error(status int, errorMsg ...string) error {
	status = ensureErrorCode(status)
	msg := http.StatusText(status)
	if len(errorMsg) > 0 {
		msg = errorMsg[0]
	}

	http.Error(r.ResponseWriter, msg, status)

	return fmt.Errorf("%v %v", status, msg)
}

func ensureErrorCode(status int) int {
	codeStr := strconv.Itoa(status)
	if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
		return status
	}

	fmt.Printf("Expected HTTP error status code. Received: %v. Returning 500 instead.\n", status)
	return 500
}

func (r *Response[T_Response]) JSON(data T_Response, status ...int) error {
	r.ResponseWriter.Header().Add("Content-Type", "application/json")

	if len(status) > 0 {
		r.ResponseWriter.WriteHeader(status[0])
	}

	if err := json.NewEncoder(r.ResponseWriter).Encode(data); err != nil {
		r.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}
