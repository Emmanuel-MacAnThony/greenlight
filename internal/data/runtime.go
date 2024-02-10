package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonvalue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonvalue)
	return []byte(quotedJSONValue), nil
}
