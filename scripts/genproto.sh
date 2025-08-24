#!/bin/bash

# Set default values
CS_PROTO_DIR="app/proto"
DEBUG=false
QUIET=true
GEN_PROTO_CODE=true
GEN_PROTO_IDS=true
CS_PROTO_FILE=""

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [PROTO_FILE]"
    echo "Options:"
    echo "  --all                Generate all proto related code (default)"
    echo "  --ids-only           Only generate protocol IDs"
    echo "  --help               Show this help message"
    echo "  PROTO_FILE           Optional path to specific proto file (absolute or relative)"
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
        --proto-file=*)
            CS_PROTO_FILE="${arg#*=}"
            # Convert to absolute path if relative
            CS_PROTO_FILE=$(realpath "$CS_PROTO_FILE")
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            if [ -z "$CS_PROTO_FILE" ]; then
                CS_PROTO_FILE="$arg"
                # Convert to absolute path if relative
                CS_PROTO_FILE=$(realpath "$CS_PROTO_FILE")
            else
                echo "Unknown option or multiple paths: $arg"
                show_usage
                exit 1
            fi
            ;;
    esac
done

# After parsing
if [ -n "$CS_PROTO_FILE" ]; then
    if [ -d "$CS_PROTO_FILE" ]; then
        CUSTOM_CS_PROTO_DIR="$CS_PROTO_FILE"
        CS_PROTO_FILE=""  # Clear since it's a dir
    else
        # It's a file, keep CS_PROTO_FILE
        :
    fi
fi

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

# Function to generate protobuf
generate_protobuf() {
    # If CUSTOM_PROTO_DIR is set, use it, otherwise use CS_PROTO_DIR
    local target_dir="${CUSTOM_CS_PROTO_DIR:-$CS_PROTO_DIR}"
    
    if [ -z "$CUSTOM_CS_PROTO_DIR" ] && [ -z "$CS_PROTO_FILE" ]; then
        echo "Deleting existing proto files..."
        find $CS_PROTO_DIR/pb -type f -not -path "*/\.*" -delete
    fi
    # For custom dir, perhaps clean specific generated files, but skip for safety
    
    echo "Generating protobuf code..."
    if [ -n "$CS_PROTO_FILE" ]; then
        protoc \
            --proto_path=. \
            --proto_path=$GOPATH/bin \
            --proto_path=$GOPATH/pkg/mod \
            --proto_path=./vendor/github.com/asynkron/protoactor-go/actor \
            --proto_path="$(dirname "$CS_PROTO_FILE")" \
            --go_out=$CS_PROTO_DIR "$CS_PROTO_FILE"
    else
        find "$target_dir" -name "*.proto" -type f | xargs -I{} protoc \
            --proto_path=. \
            --proto_path=$GOPATH/bin \
            --proto_path=$GOPATH/pkg/mod \
            --proto_path=./vendor/github.com/asynkron/protoactor-go/actor \
            --proto_path="$target_dir" \
            --go_out=$CS_PROTO_DIR {}
    fi
}

# Function to clean protocol ID files
clean_proto_ids() {
    echo "Deleting old protocol ID files..."
    find $CS_PROTO_DIR/pb -name "*_protocol_ids.go" -delete
    find $CS_PROTO_DIR/pb -name "protocol_ids.go" -delete
}

# Function to clean glue code files
clean_glue_code() {
    echo "Deleting old glue code files..."
    find $CS_PROTO_DIR/pb -name "*_request_glue.go" -delete
    find $CS_PROTO_DIR/pb -name "*_notify_glue.go" -delete
}

# Main execution logic
if [ "$GEN_PROTO_IDS" = false ]; then
    # Only generate glue code (GenGlueCode)
    if [ -z "$CS_PROTO_FILE" ]; then
        clean_glue_code
    fi
    echo "Generating glue code only..."
    go run tools/gotools/genproto/main.go --proto_dir="${CUSTOM_CS_PROTO_DIR:-$CS_PROTO_DIR}" --gen_proto_ids=false $QUIET_OPT $DEBUG_OPT ${CS_PROTO_FILE:+--proto-file=$CS_PROTO_FILE}
elif [ "$GEN_PROTO_CODE" = false ]; then
    # Only generate protocol IDs (GenProtoID)
    if [ -z "$CS_PROTO_FILE" ]; then
        clean_proto_ids
    fi
    echo "Generating protocol IDs only..."
    go run tools/gotools/genproto/main.go --proto_dir="${CUSTOM_CS_PROTO_DIR:-$CS_PROTO_DIR}" --gen_proto_code=false $QUIET_OPT $DEBUG_OPT ${CS_PROTO_FILE:+--proto-file=$CS_PROTO_FILE}
else
    # Generate all (GenProto or GenProtoDebug)
    generate_protobuf
    echo "Generating protocol IDs and glue code..."
    go run tools/gotools/genproto/main.go --proto_dir="${CUSTOM_CS_PROTO_DIR:-$CS_PROTO_DIR}" $QUIET_OPT $DEBUG_OPT ${CS_PROTO_FILE:+--proto-file=$CS_PROTO_FILE}
fi

echo "Proto generation completed."
