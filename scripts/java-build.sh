#!/bin/sh

echo "STANDARD JAVA BUILD (gitserver)"
#export GRADLE_OPTS="-Dorg.gradle.daemon=false"
#env


# java is being weird and does not use ${HOME}
export _JAVA_OPTIONS=-Duser.home=${HOME}

# gradle install
unset GRADLE_OPTS
# env
export _JAVA_OPTIONS=-Duser.home=${HOME}

echo "Home directory: ${HOME}"

# clear the cache and all stuff left flying around by
# any previous builds
# prefer to delete .gradle/ but that is bad if it
# is mistakenly run on a users' machine
rm -rf ${HOME}/.gradle/caches

export ORG_GRADLE_PROJECT_mavenRepoUser=${GE_USER_EMAIL}
export ORG_GRADLE_PROJECT_mavenRepoToken=${GE_USER_TOKEN}

# stop the annoying and unnecessary gradle "daemon"
gradle --stop

gradle build || exit 10

mkdir dist/java || exit 10
cp -rv build/libs/ dist/java || exit 10

gradle distTar || exit 10
mv build/distributions/*.tar dist/java/${PROJECT_NAME}.tar || exit 10

# stop the annoying and unnecessary gradle "daemon"
gradle --stop

echo "java compile done."
