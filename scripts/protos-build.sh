#!/bin/sh

echo "STANDARD PROTOS BUILD (gitserver)"
if [ "${FORKED}" != "true" ] && [ -x /tmp/protos-build.sh ]; then
    export FORKED=true
    . /tmp/protos-build.sh
    exit 0
fi
#env
mkdir -p build/protos/golang/src
#( cd protos && /usr/bin/protoc --go_out=plugins=grpc:../build/protos/golang/src `find -name '*.proto'` ) || exit 10
( /usr/local/bin/protorenderer-client -compile -registry=registry `find protos/ -name '*.proto'` ) || exit 10
# auto submit (but no error check)
( /usr/local/bin/protorenderer-client -repository_id=${REPOSITORY_ID} -registry=registry `find protos/ -name '*.proto'` )
echo "protos compile done."
