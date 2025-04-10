module github.com/SpirentOrion/luddite/v3

go 1.24.1

require (
	github.com/K-Phoen/negotiation v0.0.0-20160529191006-5f2c7e65d11c
	github.com/dimfeld/httptreemux v5.0.1+incompatible
	github.com/gorilla/schema v1.4.1
	github.com/opentracing/basictracer-go v1.1.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.22.0
	github.com/rs/cors v1.11.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	golang.org/x/net v0.39.0
	golang.org/x/tools v0.32.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.63.0 // indirect
	github.com/prometheus/procfs v0.16.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	golang.org/x/mod => golang.org/x/mod v0.24.0
	golang.org/x/net => golang.org/x/net v0.39.0
	golang.org/x/sync => golang.org/x/sync v0.13.0
	golang.org/x/sys => golang.org/x/sys v0.31.0
	golang.org/x/text => golang.org/x/text v0.24.0
	golang.org/x/time => golang.org/x/time v0.11.0
	golang.org/x/tools => golang.org/x/tools v0.32.0
)
