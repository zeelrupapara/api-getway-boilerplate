module gitlab.com/flexgrewtechnologies/greenlync-api-gateway

go 1.21

require (
	// Web Framework
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/gofiber/websocket/v2 v2.2.1
	
	// Authentication & Authorization
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/casbin/casbin/v2 v2.82.0
	github.com/casbin/gorm-adapter/v3 v3.20.0
	golang.org/x/crypto v0.17.0
	
	// Database
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	
	// Cache & Session Management
	github.com/go-redis/redis/v8 v8.11.5
	
	// Message Queue
	github.com/nats-io/nats.go v1.31.0
	
	// Configuration
	github.com/spf13/viper v1.18.2
	
	// Logging
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.26.0
	
	// Utilities
	github.com/google/uuid v1.6.0
	github.com/go-playground/validator/v10 v10.16.0
	
	// Monitoring & Observability
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	github.com/prometheus/client_golang v1.17.0
	
	// gRPC (for microservice communication)
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	
	// Testing
	github.com/stretchr/testify v1.8.4
	github.com/golang/mock v1.6.0
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fasthttp/websocket v1.5.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)