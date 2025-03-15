package network

import "testing"

func TestCodecServerEncodeClientDecode(t *testing.T) {
	sCodec := new(Codec)
	pack, err := sCodec.Encode([]byte("hello world"), 0, 1)
	if err != nil {
		t.Fatal(err)
	}

	cCodec := new(ClientCodec)
	msgList, err := cCodec.Decode(pack.Data(), func(pid uint32) bool { return false })
	if err != nil {
		t.Fatal(err)
	}

	msg := msgList[0]

	if msg.Pid != 1 || msg.Seq != 0 || string(msg.Data) != "hello world" {
		t.Fatalf("expected pid: 1, seq: 0, data: hello world, got: pid: %d, seq: %d, data: %s", msg.Pid, msg.Seq, string(msg.Data))
	}
}

func TestCodecClientEncodeServerDecode(t *testing.T) {
	cCodec := new(ClientCodec)
	pack := cCodec.Encode([]byte("hello world"), 0, 1)

	sCodec := new(Codec)
	msgList, err := sCodec.Decode(pack.Data())
	if err != nil {
		t.Fatal(err)
	}
	if len(msgList) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgList))
	}
	msg := msgList[0]
	if msg.Pid != 1 || msg.Seq != 0 || string(msg.Data) != "hello world" {
		t.Fatalf("expected pid: 1, seq: 0, data: hello world, got: pid: %d, seq: %d, data: %s", msg.Pid, msg.Seq, string(msg.Data))
	}
}
