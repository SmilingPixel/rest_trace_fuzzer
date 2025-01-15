# You should execute `go install github.com/google/gnostic/cmd/protoc-gen-openapi` before running this script.
# Annotate your proto file according to https://github.com/google/gnostic/issues/412
# By default, the output directory is the current directory, the output file is `openapi.yaml`

protoc \
    demo.proto \
    -I . \
    -I=./include \
    --openapi_out=.