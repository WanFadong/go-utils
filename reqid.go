package utils

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/qiniu/xlog.v1"
	"qbox.us/errors"
)

var (
	errInvalidArgs = errors.New("invalid args")
)

func DecodeReqid(xl *xlog.Logger, reqid string) (pid uint32, unixNano int64, err error) {
	b, err := base64.URLEncoding.DecodeString(reqid)
	if err != nil {
		xl.Error("invalid args: decode failed=>", err)
		err = errInvalidArgs
		return
	}
	if len(b) != 12 {
		xl.Error("invalid args: len(b) != 12")
		err = errInvalidArgs
		return
	}
	pid = uint32(binary.LittleEndian.Uint32(b[:4]))
	unixNano = int64(binary.LittleEndian.Uint64(b[4:]))
	return
}
