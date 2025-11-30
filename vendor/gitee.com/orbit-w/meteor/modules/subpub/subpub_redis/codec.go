package subpub_redis

import (
	"errors"

	jsoniter "github.com/json-iterator/go"
	proto "google.golang.org/protobuf/proto"
)

var (
	jsonAPI = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	CodecJson = iota
	CodecProto3
	CodecString
)

func encode(name int, sender string, pid int64, v any) ([]byte, error) {
	var (
		err  error
		body []byte
	)

	switch name {
	case CodecJson:
		body, err = jsonAPI.Marshal(v)
		if err != nil {
			return nil, err
		}
	case CodecProto3:
		pbMsg, ok := v.(proto.Message)
		if !ok {
			return nil, errors.New("")
		}
		body, err = proto.Marshal(pbMsg)
		if err != nil {
			return nil, err
		}
	case CodecString:
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("v is not string")
		}
		body = []byte(str)
	}

	msg := &PubMessage{
		Sender: sender,
		Pid:    pid,
		Data:   body,
	}
	return proto.Marshal(msg)
}
