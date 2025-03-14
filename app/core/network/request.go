package network

import "github.com/orbit-w/meteor/modules/net/packet"

type ClientRequest struct {
	upSeq uint32
	pid   uint32
	data  []byte

	session *Session
}

func NewClientRequest(upSeq uint32, pid uint32, data []byte, session *Session) *ClientRequest {
	return &ClientRequest{
		upSeq:   upSeq,
		pid:     pid,
		data:    data,
		session: session,
	}
}

func (r *ClientRequest) Response(data []byte, pid uint32) error {
	pack, err := codec.Encode(data, r.upSeq, pid)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return r.session.Send(pack.Data())
}

func (r *ClientRequest) ResponseBatch(msgs []Message) error {
	pack, err := codec.EncodeBatch(r.upSeq, msgs)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return r.session.Send(pack.Data())
}
