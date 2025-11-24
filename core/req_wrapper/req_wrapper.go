package reqwrapper

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"google.golang.org/protobuf/proto"
)

type RequestWrapper struct {
	Pid  uint32
	Refs []*mme.EntityRef
}

func NewRequestWrapper(pid uint32) *RequestWrapper {
	return &RequestWrapper{
		Pid: pid,
	}
}

func (qw *RequestWrapper) Unmarshal(req proto.Message, data []byte) error {
	if err := proto.Unmarshal(data, req); err != nil {
		return err
	}

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
