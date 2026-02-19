module github.com/poi2/building-a-schema-first-dynamic-validation-system/services/isr

go 1.21

require (
	connectrpc.com/connect v1.17.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go v0.0.0
	golang.org/x/net v0.33.0
	google.golang.org/protobuf v1.35.2
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.35.2-20241127180247-a33202765966.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/poi2/building-a-schema-first-dynamic-validation-system/pkg/gen/go => ../../pkg/gen/go
