package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func DecodeJSONStruct(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return errors.New(formatJSONErrorMessage(err.Error()))
	}
	return nil
}

func formatJSONErrorMessage(errMsg string) string {
	if strings.Contains(errMsg, "unknown field") {
		fieldStart := strings.Index(errMsg, "\"")
		if fieldStart != -1 {
			fieldEnd := strings.Index(errMsg[fieldStart+1:], "\"")
			if fieldEnd != -1 {
				filedName := errMsg[fieldStart+1 : fieldStart+1+fieldEnd]
				return "Unknown field: " + filedName
			}
		}
	}
	return "Invalid json format"
}
