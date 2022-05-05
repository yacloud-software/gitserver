.PHONY:	client server hooks gitcredentials goversion
all:	client server hooks gitcredentials goversion
gitcredentials:
	cd gitcredentials && go install gitserver-credentials.go `ls -1 *.go|grep -v gitserver-credentials`
client:
	cd client && go install gitserver-client.go `ls -1 *.go|grep -v gitserver-client.go`
server:
	cd server && go install gitserver-server.go `ls -1 *.go|grep -v gitserver-server`
hooks:
	cd hooks && go install git-hook.go `ls -1 *.go|grep -v git-hook.go`
goversion:
	cd goversion && go install go-version.go `ls -1 *.go|grep -v go-version.go`
