#!/bin/sh

echo "Cleaning repository for new build..."
[ -f errorfile ] && rm errorfile
[ -d dist ] && rm -rf dist
[ -d build ] && rm -rf build
mkdir dist
mkdir build

