package reqresp

type Request struct {
	ID   string
	Data any
	Seq  uint32
	PID  uint32
}
