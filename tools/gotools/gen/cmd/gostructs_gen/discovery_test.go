package cmd

import (
	"fmt"
	"testing"
)

func TestFindDescriptorProto(t *testing.T) {
	protobufIncludePath := findDescriptorProto()
	fmt.Println(protobufIncludePath)
}

func TestGetProtobufIncludePaths(t *testing.T) {
	paths := getProtobufIncludePaths()
	fmt.Println(paths)
}
