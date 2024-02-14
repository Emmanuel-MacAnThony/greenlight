package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonvalue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonvalue)
	return []byte(quotedJSONValue), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {

	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i)

	return nil

}
