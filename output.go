package utils

import (
	"encoding/json"
	"fmt"
)

func GetPrettyJson(r interface{}) (str string, err error) {
	b, err := json.MarshalIndent(r, "", "    ")
	str = string(b)
	return
}

func GetJson(r interface{}) (str string, err error) {
	b, err := json.Marshal(r)
	str = string(b)
	return
}

func OutputPrettyJson(r interface{}) {
	str, err := GetPrettyJson(r)
	PanicIfError(err)
	fmt.Println(str)
}
