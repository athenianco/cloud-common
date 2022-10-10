module github.com/athenianco/cloud-common

go 1.16

require (
	cloud.google.com/go v0.101.1 // indirect
	cloud.google.com/go/kms v1.4.0
	cloud.google.com/go/pubsub v1.21.1
	cloud.google.com/go/storage v1.22.0
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/cli v20.10.16+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.16+incompatible // indirect
	github.com/getsentry/sentry-go v0.14.0
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/jackc/pgx/v4 v4.16.1
	github.com/lib/pq v1.10.6
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/ory/dockertest/v3 v3.8.1
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.37.0
	github.com/rs/zerolog v1.26.1
	github.com/slack-go/slack v0.10.3
	github.com/stretchr/testify v1.8.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	google.golang.org/api v0.80.0 // indirect
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3
	google.golang.org/grpc v1.46.2 // indirect
)

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.2
	github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.6
	github.com/containerd/imgcrypt => github.com/containerd/imgcrypt v1.1.4
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.8.1
	github.com/coreos/etcd => go.etcd.io/etcd/v3 v3.5.5
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/docker/cli => github.com/docker/cli v20.10.16+incompatible
	github.com/docker/distribution => github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.16+incompatible
	github.com/emicklei/go-restful => github.com/emicklei/go-restful/v3 v3.8.0
	github.com/gobuffalo/packr => github.com/gobuffalo/packr/v2 v2.3.2
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/handlers => github.com/gorilla/handlers v1.3.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.1
	github.com/jackc/pgproto3 => github.com/jackc/pgproto3/v2 v2.1.1
	github.com/kataras/iris => github.com/kataras/iris/v12 v12.2.0-alpha8
	github.com/microcosm-cc/bluemonday => github.com/microcosm-cc/bluemonday v1.0.16
	github.com/miekg/dns => github.com/miekg/dns v1.1.25
	github.com/nats-io/jwt => github.com/nats-io/jwt/v2 v2.0.1
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.13.0
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.34.0
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be
	golang.org/x/net => golang.org/x/net v0.0.0-20220927171203-f486391704dc
	golang.org/x/text => golang.org/x/text v0.3.7
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1
	k8s.io/kubernetes => k8s.io/kubernetes v1.18.6
)
