module github.com/aserto-dev/self-decision-logger

go 1.22.0

toolchain go1.22.5

// replace github.com/aserto-dev/go-decision-logs => ../go-decision-logs

require (
	github.com/aserto-dev/go-aserto v0.32.2-0.20241111205103-d469387b8361
	github.com/aserto-dev/go-authorizer v0.20.11
	github.com/aserto-dev/go-decision-logs v0.1.2
	github.com/aserto-dev/topaz v0.32.33
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/google/uuid v1.6.0
	github.com/nats-io/nats-server/v2 v2.10.22
	github.com/nats-io/nats.go v1.37.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/sync v0.9.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)

require (
	github.com/aserto-dev/header v0.0.8 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.23.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/nats-io/jwt/v2 v2.5.8 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/samber/lo v1.47.0 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
