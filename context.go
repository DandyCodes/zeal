package zeal

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Ctx[T_Response any] struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (c *Ctx[T_Response]) JSON(data T_Response, status ...int) {
	c.ResponseWriter.Header().Add("Content-Type", "application/json")
	if len(status) > 0 {
		firstStatus := status[0]
		c.ResponseWriter.WriteHeader(firstStatus)
	}
	if err := json.NewEncoder(c.ResponseWriter).Encode(data); err != nil {
		c.ResponseWriter.WriteHeader(http.StatusInternalServerError)
	}
}

func (c Ctx[T_Response]) Status(status int) {
	c.ResponseWriter.WriteHeader(status)
}

func (c Ctx[T_Response]) Error(status int, errorMsg ...string) {
	if len(errorMsg) > 0 {
		http.Error(c.ResponseWriter, errorMsg[0], ensureErrorCode(status))
	} else {
		http.Error(c.ResponseWriter, http.StatusText(status), ensureErrorCode(status))
	}
}

func ensureErrorCode(status int) int {
	codeStr := strconv.Itoa(status)
	if len(codeStr) == 3 && (codeStr[0] == '4' || codeStr[0] == '5') {
		return status
	}
	return 500
}
