module github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr

go 1.24.0

require (
	connectrpc.com/connect v1.19.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go v0.0.0
	golang.org/x/net v0.37.0
	google.golang.org/protobuf v1.36.9
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1 // indirect
	buf.build/go/protovalidate v1.0.0 // indirect
	cel.dev/expr v0.24.0 // indirect
	connectrpc.com/validate v0.6.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250922171735-9219d122eba9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250922171735-9219d122eba9 // indirect
)

replace github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go => ../../pkg/gen/go
