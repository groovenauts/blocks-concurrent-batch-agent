package models

import (
	"encoding/json"

	validator "gopkg.in/go-playground/validator.v9"
)

var Validators = map[string]validator.Func{
	"json":        isValidJSON,
	"jsonstr2str": isValidJSONStr2Str,
}

func isValidJSON(fl validator.FieldLevel) bool {
	var dest interface{}
	err := json.Unmarshal([]byte(fl.Field().String()), &dest)
	return err == nil
}

func isValidJSONStr2Str(fl validator.FieldLevel) bool {
	var dest map[string]string
	str := fl.Field().String()
	err := json.Unmarshal([]byte(str), &dest)
	return err == nil
}
