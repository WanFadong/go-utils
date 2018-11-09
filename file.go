package utils

import (
	"os"
	"path/filepath"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

// 打开或新建一个文件，用于读写（使用追加的方式）
// 同时会创建相应的目录
// 返回是否是新建的文件
func OpenOrCreateFile(filename string) (file *os.File, fileExists bool, err error) {
	fileExists, err = IsFileExists(filename)
	if err != nil {
		return
	}

	var flag int
	if fileExists {
		flag = os.O_RDWR | os.O_APPEND
	} else {
		if err = os.MkdirAll(filepath.Dir(filename), 0775); err != nil {
			return
		}
		flag = os.O_RDWR | os.O_CREATE | os.O_EXCL | os.O_TRUNC
	}
	file, err = os.OpenFile(filename, flag, 0666)
	if err != nil {
		log.Errorf("Failed to open file, fileExists: %v, err: %v", fileExists, err)
		return
	}
	return
}

func CreateOrRemoveFile(xl *xlog.Logger, filename string) (file *os.File, err error) {
	fileExists, err := IsFileExists(filename)
	if err != nil {
		return
	}

	if fileExists {
		err = os.Remove(filename)
		if err != nil {
			return
		}
	}

	if err = os.MkdirAll(filepath.Dir(filename), 0775); err != nil {
		return
	}
	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL | os.O_TRUNC
	file, err = os.OpenFile(filename, flag, 0666)
	if err != nil {
		log.Errorf("Failed to open file, fileExists: %v, err: %v", fileExists, err)
		return
	}
	return
}

func IsFileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
