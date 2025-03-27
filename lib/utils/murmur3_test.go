package utils

import (
	"fmt"
	"testing"
)

func TestMurmurHash3_x86_32(t *testing.T) {
	// Test internal consistency
	cases := []string{
		"",
		"hello",
		"hello, world!",
		"The quick brown fox jumps over the lazy dog",
	}

	// Test that the same input always produces the same output
	for i, input := range cases {
		t.Run(fmt.Sprintf("Consistency test %d", i), func(t *testing.T) {
			first := MurmurHash3_x86_32([]byte(input), 0)
			second := MurmurHash3_x86_32([]byte(input), 0)
			if first != second {
				t.Errorf("MurmurHash3_x86_32 not consistent for input %q: %x != %x", input, first, second)
			}
		})
	}

	// Test that different inputs produce different outputs (hash property)
	for i := 0; i < len(cases); i++ {
		for j := i + 1; j < len(cases); j++ {
			t.Run(fmt.Sprintf("Difference test %d vs %d", i, j), func(t *testing.T) {
				hash1 := MurmurHash3_x86_32([]byte(cases[i]), 0)
				hash2 := MurmurHash3_x86_32([]byte(cases[j]), 0)
				if hash1 == hash2 {
					t.Errorf("Hash collision between %q and %q: both produced %x", cases[i], cases[j], hash1)
				}
			})
		}
	}

	// Test that different seeds produce different outputs for the same input
	for i, input := range cases {
		t.Run(fmt.Sprintf("Seed test %d", i), func(t *testing.T) {
			hash1 := MurmurHash3_x86_32([]byte(input), 0)
			hash2 := MurmurHash3_x86_32([]byte(input), 42)
			if hash1 == hash2 {
				t.Errorf("Seed had no effect on hash for input %q: both produced %x", input, hash1)
			}
		})
	}
}

func TestMurmurHash3_x64_128(t *testing.T) {
	// Test internal consistency
	cases := []string{
		"",
		"hello",
		"hello, world!",
		"The quick brown fox jumps over the lazy dog",
	}

	// Test that the same input always produces the same output
	for i, input := range cases {
		t.Run(fmt.Sprintf("Consistency test %d", i), func(t *testing.T) {
			first := MurmurHash3_x64_128([]byte(input), 0)
			second := MurmurHash3_x64_128([]byte(input), 0)
			if first != second {
				t.Errorf("MurmurHash3_x64_128 not consistent for input %q: %v != %v", input, first, second)
			}
		})
	}

	// Test that different inputs produce different outputs (hash property)
	for i := 0; i < len(cases); i++ {
		for j := i + 1; j < len(cases); j++ {
			t.Run(fmt.Sprintf("Difference test %d vs %d", i, j), func(t *testing.T) {
				hash1 := MurmurHash3_x64_128([]byte(cases[i]), 0)
				hash2 := MurmurHash3_x64_128([]byte(cases[j]), 0)
				if hash1 == hash2 {
					t.Errorf("Hash collision between %q and %q: both produced %v", cases[i], cases[j], hash1)
				}
			})
		}
	}

	// Test that different seeds produce different outputs for the same input
	for i, input := range cases {
		t.Run(fmt.Sprintf("Seed test %d", i), func(t *testing.T) {
			hash1 := MurmurHash3_x64_128([]byte(input), 0)
			hash2 := MurmurHash3_x64_128([]byte(input), 42)
			if hash1 == hash2 {
				t.Errorf("Seed had no effect on hash for input %q: both produced %v", input, hash1)
			}
		})
	}
}

func TestStringHash32(t *testing.T) {
	// Test that StringHash32 gives the same results as MurmurHash3_x86_32
	tests := []struct {
		input string
		seed  uint32
	}{
		{"", 0},
		{"hello", 0},
		{"hello, world!", 0},
		{"hello, world!", 42},
		{"The quick brown fox jumps over the lazy dog", 0},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			expected := MurmurHash3_x86_32([]byte(test.input), test.seed)
			got := StringHash32(test.input, test.seed)
			if got != expected {
				t.Errorf("StringHash32(%q, %d) = 0x%x, want 0x%x", test.input, test.seed, got, expected)
			}
		})
	}
}

func TestStringHash128(t *testing.T) {
	// Test that StringHash128 gives the same results as MurmurHash3_x64_128
	tests := []struct {
		input string
		seed  uint64
	}{
		{"", 0},
		{"hello", 0},
		{"hello, world!", 0},
		{"hello, world!", 42},
		{"The quick brown fox jumps over the lazy dog", 0},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			expected := MurmurHash3_x64_128([]byte(test.input), test.seed)
			got := StringHash128(test.input, test.seed)
			if got != expected {
				t.Errorf("StringHash128(%q, %d) = %v, want %v", test.input, test.seed, got, expected)
			}
		})
	}
}

func BenchmarkMurmurHash3_x86_32(b *testing.B) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MurmurHash3_x86_32(data, 0)
	}
}

func BenchmarkMurmurHash3_x64_128(b *testing.B) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MurmurHash3_x64_128(data, 0)
	}
}

func BenchmarkStringHash32(b *testing.B) {
	s := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringHash32(s, 0)
	}
}

func BenchmarkStringHash128(b *testing.B) {
	s := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringHash128(s, 0)
	}
}

func TestHashProtoMessage(t *testing.T) {
	messageName := "Core-Request_SearchBook_Rsp"
	hash := HashProtoMessage(messageName)
	fmt.Println(fmt.Printf("Hash of %s: 0x%08x\n", messageName, hash))
}
