package utils

import (
	"encoding/base64"
	"encoding/binary"
	"github.com/pkg/errors"
	"time"
)

var (
	errInvalidArgs = errors.New("invalid args")
)

func DecodeReqid(reqid string) (pid uint32, t time.Time, err error) {
	b, err := base64.URLEncoding.DecodeString(reqid)
	if err != nil {
		err = errors.WithMessage(errInvalidArgs, "base64 decode failed")
		return
	}
	if len(b) != 12 {
		err = errors.WithMessage(errInvalidArgs, "len(b) != 12")
		return
	}
	pid = uint32(binary.LittleEndian.Uint32(b[:4]))
	unixNano := int64(binary.LittleEndian.Uint64(b[4:]))
	t = time.Unix(unixNano/1e9, unixNano%1e9)
	return
}
