#!/bin/sh

# bundle up the dist
echo Building dist...

TARDIRS="dist/"
if [ -d configs ]; then
    TARDIRS="$TARDIRS configs/"
fi
if [ -d scripts ]; then
    TARDIRS="$TARDIRS scripts/"
fi
if [ -d templates ]; then
    TARDIRS="$TARDIRS templates/"
fi

if [ -d extra ]; then
    TARDIRS="$TARDIRS extra/"
fi

if [ -d lib ]; then
    TARDIRS="$TARDIRS lib/"
fi


tar -cf dist.tar ${TARDIRS} || fatal_error tar failed
mv dist.tar dist/ || fatal_error tar move failed


# we're not on a build server..
if [ -z "${BUILD_NUMBER}" ]; then
    echo no build number - not submitting
    exit 0
fi

#echo "CTX environment variable: \"${GE_CTX}\""
#echo "Whoami:"
#cwhoami

# we are on a build server, so submit it:
build-repo-client -repository_id=${REPOSITORY_ID} -branch=${GIT_BRANCH} -build=${BUILD_NUMBER} -commitid=${COMMIT_ID} -commitmsg="commit msg unknown" -repository=${BUILD_ARTEFACT} || exit 10

echo Completed build ${BUILD_NUMBER}

