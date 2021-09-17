module github.com/tinkerbell/boots

go 1.16

require (
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/gammazero/workerpool v0.0.0-20200311205957-7b00833861c6
	github.com/go-logr/zapr v0.4.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/packethost/cacher v0.0.0-20200825140532-0b62e6726807
	github.com/packethost/dhcp4-go v0.0.0-20190402165401-39c137f31ad3
	github.com/packethost/pkg v0.0.0-20210325161133-868299771ae0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35
	github.com/stretchr/testify v1.7.0
	github.com/tinkerbell/tftp-go v0.0.0-20200825172122-d9200358b6cd
	github.com/tinkerbell/tink v0.0.0-20201109122352-0e8e57332303
	github.com/tobert/otel-init-go v0.0.0-20210804183951-47b83e855e3b
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/trace v1.0.0-RC2
	go.uber.org/multierr v1.6.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/tools v0.1.5
	google.golang.org/grpc v1.40.0
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	knative.dev/pkg v0.0.0-20210827184538-2bd91f75571c
	sigs.k8s.io/controller-runtime v0.9.6
)

replace github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35 => github.com/packethost/xff v0.0.0-20190305172552-d3e9190c41b3
