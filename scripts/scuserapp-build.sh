#!/bin/sh
fatal_error() {
    toilet -f mono9 FAILED
 echo $@
 exit 10
}

if [ -z "${BUILD_NUMBER}" ]; then
    export BUILD_NUMBER=1
fi

mkdir -p dist/userapp >/dev/null 2>&1

############################# create the buildnumber header files

echo ${BUILD_NUMBER} >dist/build_number_compiled.txt
cat >generated/build_version.h <<__EOF
#ifndef __CNWBUILD_VERSION_H__
#define __CNWBUILD_VERSION_H__ 1

#define CNW_BUILD_TIMESTAMP `date +%s`
#define CNW_BUILD_VERSION   ${BUILD_NUMBER}
const char *const get_cnw_build_info();
#endif

__EOF

cat >generated/build_version.c <<__EOF
const char *const get_cnw_build_info() {
 return "`date`";
}
__EOF

############### make the code

make BUILDNUMBER=${BUILD_NUMBER} || fatal_error Build Failed

#################### copy to dist
cp -v src/*.elf src/*.hex src/*.bin dist/userapp/||fatal_error failed to copy app

#################### add md5

md5sum dist/userapp/app1.bin >dist/userapp/app1.bin.md5 || fatal_error failed to create md5
md5sum dist/userapp/app2.bin >dist/userapp/app2.bin.md5 || fatal_error failed to create md5

