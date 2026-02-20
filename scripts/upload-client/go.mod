module github.com/poi2/building-a-schema-first-dynamic-validation-system/scripts/upload-client

go 1.21

require (
	connectrpc.com/connect v1.17.0
	github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go v0.0.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.35.2-20241127180247-a33202765966.1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go => ../../pkg/gen/go
