package servicezone_behavior

import (
	"fmt"

	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"gitee.com/orbit-w/orbit/pkg/proto/pb"
	"google.golang.org/protobuf/proto"
)

type RequestWrapper struct {
	Pid  uint32
	Req  proto.Message
	Refs []*mme.EntityRef
}

func NewRequestWrapper(pid uint32) *RequestWrapper {
	return &RequestWrapper{
		Pid: pid,
	}
}

func (qw *RequestWrapper) Unmarshal(data []byte) error {
	reqFactory := pb.GetPBFactory(qw.Pid)
	if reqFactory == nil {
		return fmt.Errorf("request factory not found for pid: %d", qw.Pid)
	}
	req := reqFactory()

	if err := proto.Unmarshal(data, req); err != nil {
		return err
	}
	qw.Req = req

	refFactory := pb.GetExtract(qw.Pid)
	if refFactory != nil {
		var err error
		qw.Refs, err = refFactory(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RequestWrapper) GetPid() uint32 {
	return r.Pid
}

func (r *RequestWrapper) GetRefs() []*mme.EntityRef {
	return r.Refs
}

func (r *RequestWrapper) GetReq() proto.Message {
	return r.Req
}
