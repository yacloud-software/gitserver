module golang.conradwood.net/gitserver

go 1.18

require (
	golang.conradwood.net/apis/artefact v1.1.2675
	golang.conradwood.net/apis/auth v1.1.2691
	golang.conradwood.net/apis/buildrepo v1.1.2675
	golang.conradwood.net/apis/common v1.1.2691
	golang.conradwood.net/apis/gitbuilder v1.1.2675
	golang.conradwood.net/apis/gitserver v1.1.1702
	golang.conradwood.net/apis/objectauth v1.1.2675
	golang.conradwood.net/apis/slackgateway v1.1.2675
	golang.conradwood.net/go-easyops v0.1.22616
	golang.org/x/sys v0.14.0
	google.golang.org/grpc v1.59.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.conradwood.net/apis/autodeployer v1.1.2675 // indirect
	golang.conradwood.net/apis/deploymonkey v1.1.2675 // indirect
	golang.conradwood.net/apis/echoservice v1.1.2675 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.2675 // indirect
	golang.conradwood.net/apis/framework v1.1.2675 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2691 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.2675 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2675 // indirect
	golang.conradwood.net/apis/registry v1.1.2675 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.yacloud.eu/apis/fscache v1.1.2675 // indirect
	golang.yacloud.eu/apis/session v1.1.2691 // indirect
	golang.yacloud.eu/apis/urlcacher v1.1.2675 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace golang.conradwood.net/apis/gitserver => ../../golang.conradwood.net/apis/gitserver
