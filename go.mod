module github.com/aserto-dev/self-decision-logger

go 1.22.11

toolchain go1.23.5

// replace github.com/aserto-dev/go-decision-logs => ../go-decision-logs

require (
	github.com/aserto-dev/go-aserto v0.33.6
	github.com/aserto-dev/go-authorizer v0.20.13
	github.com/aserto-dev/go-decision-logs v0.1.4
	github.com/google/uuid v1.6.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nats-io/nats-server/v2 v2.10.25
	github.com/nats-io/nats.go v1.38.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/sync v0.10.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.3
)

require (
	github.com/aserto-dev/header v0.0.10 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/nats-io/jwt/v2 v2.7.3 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/samber/lo v1.47.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
