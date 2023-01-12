.PHONY:	client server hooks gitcredentials goversion
all:	client server hooks gitcredentials goversion
gitcredentials:
	cd gitcredentials/client && go install ${LDFLAGS} gitcredentials-client.go `ls -1 *.go|grep -v gitcredentials-client.go`
	cd gitcredentials/server && go install ${LDFLAGS} gitcredentials-server.go `ls -1 *.go|grep -v gitcredentials-server.go`
client:
	cd client && go install ${LDFLAGS} gitserver-client.go `ls -1 *.go|grep -v gitserver-client.go`
server:
	cd server && go install ${LDFLAGS} gitserver-server.go `ls -1 *.go|grep -v gitserver-server`
hooks:
	cd hooks && go install ${LDFLAGS} git-hook.go `ls -1 *.go|grep -v git-hook.go`
goversion:
	cd goversion && go install ${LDFLAGS} go-version.go `ls -1 *.go|grep -v go-version.go`
