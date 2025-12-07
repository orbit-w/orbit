package netutils

import "testing"

func Test_GetLocalIPv4(t *testing.T) {
	ip, err := GetLocalIPv4()
	if err != nil {
		t.Fatalf("get local ipv4 failed: %v", err)
	}
	t.Logf("get local ipv4: %s", ip)
}

func Test_GetPublicIPv4(t *testing.T) {
	ip, err := GetPublicIPv4()
	if err != nil {
		t.Fatalf("get public ipv4 failed: %v", err)
	}
	t.Logf("get public ipv4: %s", ip)
}
