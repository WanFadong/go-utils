package go_utils

import (
	"strconv"

	xlog "github.com/sirupsen/logrus"
)

func ConvertTo36(xl *xlog.Logger, str10 string) (str36 string, err error) {
	int10, err := strconv.ParseUint(str10, 10, 32)
	if err != nil {
		xl.Error(err)
		return
	}
	str36 = strconv.FormatInt(int64(int10), 36)
	return
}

func ConvertTo10(xl *xlog.Logger, str36 string) (str10 string, err error) {
	int10, err := strconv.ParseUint(str36, 36, 64)
	if err != nil {
		xl.Error(err)
		return
	}
	str10 = strconv.FormatInt(int64(int10), 10)
	return
}
