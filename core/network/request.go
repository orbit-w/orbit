package network

type ClientRequest struct {
	upSeq uint32
	pid   uint32
	in    []byte

	session *Session
}

func NewClientRequest(seq uint32, pid uint32, in []byte, session *Session) *ClientRequest {
	return &ClientRequest{
		upSeq:   seq,
		pid:     pid,
		in:      in,
		session: session,
	}
}

func (r *ClientRequest) GetPid() uint32 {
	return r.pid
}

func (r *ClientRequest) GetSeq() uint32 {
	return r.upSeq
}

func (r *ClientRequest) GetIn() []byte {
	return r.in
}

func (r *ClientRequest) GetSession() *Session {
	return r.session
}

func (r *ClientRequest) Response(data []byte, pid uint32) error {
	return r.session.SendData(data, r.upSeq, pid)
}

func (r *ClientRequest) ResponseBatch(msgs []Message) error {
	return r.session.SendMessageBatch(msgs)
}
