package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	protoDir     = flag.String("proto_dir", "app/proto", "Directory containing .proto files")
	outputDir    = flag.String("output_dir", "app/proto/pb", "Directory for generated protocol ID files")
	debugMode    = flag.Bool("debug", false, "Enable debug mode")
	quietMode    = flag.Bool("quiet", true, "Quiet mode: only show errors")
	genProtoCode = flag.Bool("gen_proto_code", true, "Generate glue code for proto messages")
	genProtoIDs  = flag.Bool("gen_proto_ids", true, "Generate protocol IDs for proto messages")
)

func main() {
	flag.Parse()

	// Call the combined gluegen tool to generate both protocol IDs and glue code
	cmd := exec.Command("go", "run", "lib/genproto/cmd/gluegen/main.go",
		"-proto_dir", *protoDir,
		"-output_dir", *outputDir)

	if *debugMode {
		cmd.Args = append(cmd.Args, "-debug")
	}

	if *quietMode {
		cmd.Args = append(cmd.Args, "-quiet")
	}

	if !*genProtoCode {
		cmd.Args = append(cmd.Args, "-gen_proto_code=false")
	}

	if !*genProtoIDs {
		cmd.Args = append(cmd.Args, "-gen_proto_ids=false")
	}

	// Set the output to our stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating protocol data: %v\n", err)
	}
}
