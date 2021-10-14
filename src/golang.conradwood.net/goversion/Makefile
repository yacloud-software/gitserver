all:
	cd client && go install go-version.go
gitlab:
	scp ${GOBIN}/go-version gitlab-runner:/usr/local/bin/
