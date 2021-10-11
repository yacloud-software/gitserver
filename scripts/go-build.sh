#!/bin/sh

# go insists on absolute path.
export GOBIN=`pwd`/dist
export DISTDIR=`pwd`/dist
export GOPATH=`pwd`
echo "GOPATH=$GOPATH"

# space seperated packages
if [ -d src ]; then
    DIRS=`ls -1 src/`
    for thi in ${DIRS}; do
	if [ -d src/$thi ]; then
	    PACKAGES="${PACKAGES} `cd src/$thi && ls -1 |grep -v vendor`"
	fi
    done
fi

# gitlab nonsense
[ -z "${BUILD_NUMBER}" ] && export BUILD_NUMBER=${CI_PIPELINE_ID}
[ -z "${PROJECT_NAME}" ] && export PROJECT_NAME=${CI_PROJECT_NAME}
[ -z "${COMMIT_ID}" ] && export COMMIT_ID=${CI_COMMIT_SHA}
[ -z "${GIT_BRANCH}" ] && export GIT_BRANCH=${CI_COMMIT_REF_NAME}


# use this to tar stuff
finishhook() {
	echo finished.
}

fatal_error() {
    toilet ERROR
    echo $@
    exit 10
}

# for each submodule in subs, build it
build_submodules() {
    if [ ! -d subs ]; then
	return
    fi
    echo "Building submodules"
    date
    for sub in subs/* ; do
	for mk in $sub/src/*; do
	    for pkgs in $mk/*; do
		echo $pkgs|grep -q '/vendor$'
		[ $? -eq 0 ] && continue
		#	    [ -f $mk/Makefile ] || continue
		echo "Building SubModule:"
		echo "  sub : $sub"
		echo "  mk  : $mk"
		echo "  pkgs: $pkgs"
		( export GOPATH=$GOPATH/$sub ; cd $pkgs && make )
	    done
	done
    done    
}

# for each in @PACKAGE, which has a makefile, compile it
buildall() {
    echo building ${GOOS}/$GOARCH
    GOBIN=${DISTDIR}/${GOOS}/${GOARCH}
    mkdir -p $GOBIN
    EFILE="`pwd`/errorfile"
    [ -f $EFILE ] && rm $EFILE
    echo PACKAGES: $PACKAGES
    for pkg in ${PACKAGES}; do
	for D in ${DIRS}; do
	    MYSRC=src/${D}/${pkg}
	    if [ -f $MYSRC/Makefile ]; then
		( cd ${MYSRC} && make all || echo $MYSRC >>$EFILE ) &
	    fi
	done
    done

    # also build submodules
    build_submodules
    
    wait
    if [ -f $EFILE ]; then
	echo "Failed packages:"
	cat $EFILE
	fatal_error compile failed
    fi
}

buildModules() {
    [ -d src ] || return
    for DIR in `ls -1 src/`; do
	echo "DIR: ${DIR}"
	MODS=`ls -1 src/${DIR}|grep -v vendor`
	for MOD in $MODS ; do
	    FMOD=${DIR}/${MOD}
	    TARFILE="${DIR}_${MOD}.tar.bz2"
	    echo "Module: \"${FMOD}\" -> ${TARFILE}"
	    tar -jcf dist/${TARFILE} -C src ${FMOD} || fatal_error failed to tar ${FMOD} into ${TARFILE}
	done
    done
    
}


echo
echo
echo "go-build.sh for ${PROJECT_NAME}, build #${BUILD_NUMBER}"
echo

# we only build for amd64 atm
export GOARCH=amd64

# this allows local builds on -dev machines
# to quickly build only a single arch
# intent is for devs to set DEVOS=[localos] permanently
# on their machine and
# the autobuild.sh will do 'The Right Thing'
if [ ! -z "${DEVOS}" ]; then
    echo "Building developer version for $DEVOS"
    GOOS=${DEVOS}
    buildall
    finishhook
    exit 0
fi
if [ ! -z "$BUILD_NUMBER" ] && [ -d src ]; then
    /usr/local/bin/go-version -repo=${PROJECT_NAME} || fatal_error go version failed
fi
#========= build linux
if [ -z "$BUILD_TARGETS" ]; then
    BUILD_TARGETS="linux"
fi
for thi in $BUILD_TARGETS; do
    export GOOS=$thi
    buildall
done


buildModules

echo
echo
echo '********** Go-VET *******'
if [ -f /tmp/vet_disabled ]; then
    echo "VET DISABLED"
else
    if [ -d src ]; then
	/opt/cnw/ctools/dev/bin/vet || fatal_error VET_ERROR
    fi
fi



finishhook

