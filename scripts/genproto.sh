#!/bin/bash

# Set default values
PROTO_DIR="app/proto"
DEBUG=false
QUIET=true
GEN_PROTO_CODE=true
GEN_PROTO_IDS=true

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --all                Generate all proto related code (default)"
    echo "  --ids-only           Only generate protocol IDs"
    echo "  --glue-only          Only generate glue code"
    echo "  --debug              Run in debug mode"
    echo "  --proto-dir=DIR      Set proto directory (default: app/proto)"
    echo "  --help               Show this help message"
}

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        --all)
            GEN_PROTO_CODE=true
            GEN_PROTO_IDS=true
            shift
            ;;
        --ids-only)
            GEN_PROTO_CODE=false
            GEN_PROTO_IDS=true
            shift
            ;;
        --glue-only)
            GEN_PROTO_CODE=true
            GEN_PROTO_IDS=false
            shift
            ;;
        --debug)
            DEBUG=true
            QUIET=false
            shift
            ;;
        --proto-dir=*)
            PROTO_DIR="${arg#*=}"
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            # Unknown option
            echo "Unknown option: $arg"
            show_usage
            exit 1
            ;;
    esac
done

# Create debug option string
DEBUG_OPT=""
if [ "$DEBUG" = true ]; then
    DEBUG_OPT="--debug"
fi

# Create quiet option string
QUIET_OPT="--quiet"
if [ "$QUIET" = false ]; then
    QUIET_OPT="--quiet=false"
fi

# Function to generate protobuf code
generate_protobuf() {
    # Delete all files in app/proto/pb
    echo "Deleting existing proto files..."
    find $PROTO_DIR/pb -type f -not -path "*/\.*" -delete
    
    # Generate protobuf code
    echo "Generating protobuf code..."
    find $PROTO_DIR -name "*.proto" -type f | xargs -I{} protoc \
        --proto_path=. \
        --proto_path=$GOPATH/bin \
        --proto_path=$GOPATH/pkg/mod \
        --proto_path=./vendor/github.com/asynkron/protoactor-go/actor \
        --go_out=$PROTO_DIR {}
}

# Function to clean protocol ID files
clean_proto_ids() {
    echo "Deleting old protocol ID files..."
    find $PROTO_DIR/pb -name "*_protocol_ids.go" -delete
    find $PROTO_DIR/pb -name "protocol_ids.go" -delete
}

# Function to clean glue code files
clean_glue_code() {
    echo "Deleting old glue code files..."
    find $PROTO_DIR/pb -name "*_request_glue.go" -delete
    find $PROTO_DIR/pb -name "*_notify_glue.go" -delete
}

# Main execution logic
if [ "$GEN_PROTO_IDS" = false ]; then
    # Only generate glue code (GenGlueCode)
    clean_glue_code
    echo "Generating glue code only..."
    go run tools/gotools/genproto/main.go --proto_dir=$PROTO_DIR --gen_proto_ids=false $QUIET_OPT $DEBUG_OPT
elif [ "$GEN_PROTO_CODE" = false ]; then
    # Only generate protocol IDs (GenProtoID)
    clean_proto_ids
    echo "Generating protocol IDs only..."
    go run tools/gotools/genproto/main.go --proto_dir=$PROTO_DIR --gen_proto_code=false $QUIET_OPT $DEBUG_OPT
else
    # Generate all (GenProto or GenProtoDebug)
    generate_protobuf
    echo "Generating protocol IDs and glue code..."
    go run tools/gotools/genproto/main.go --proto_dir=$PROTO_DIR $QUIET_OPT $DEBUG_OPT
fi

echo "Proto generation completed."
