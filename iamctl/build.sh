#!/bin/bash

# Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
#
# WSO2 Inc. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.

function showUsageAndExit() {
    echo "Insufficient or invalid options provided"
}

function detectPlatformSpecificBuild() {
    platform=$(uname -s)
    if [[ "${platform}" == "Linux" ]]; then
        platforms="linux/386/linux/i586 linux/amd64/linux/x64"
    elif [[ "${platform}" == "Darwin" ]]; then
        platforms="darwin/amd64/macosx/x64"
    else
        platforms="windows/386/windows/i586 windows/amd64/windows/x64"
    fi
}
#target="main.go"
#build_version="1.2"
#full_build="true"

while getopts :t:v:f FLAG; do
  case $FLAG in
    t)
      target=$OPTARG
      ;;
    v)
      build_version=$OPTARG
      ;;
    f)
      full_build="true"
      ;;
    \?)
      showUsageAndExit
      ;;
  esac
done

strSkipTest="test.skip"

for arg in "$@"
do
    if [ "$arg" == "$strSkipTest" ]
    then
        skipTest=true
    fi
done

if [ ! -e "$target" ]; then
  echo "Target file is needed. "
  showUsageAndExit
  exit 1
fi

if [ -z "$build_version" ]
then
  echo "Build version is needed. "
  showUsageAndExit
fi


rootPath=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
buildDir="build"
buildPath="$rootPath/${buildDir}"

echo "Cleaning build path ${buildDir}..."
rm -rf $buildPath

filename=$(basename ${target})
baseDir=$(dirname ${target})
if [ ".go" == ${filename:(-3)} ]
then
    filename=${filename%.go}
fi

#platforms="darwin/amd64 freebsd/386 freebsd/amd64 freebsd/arm linux/386 linux/amd64 linux/arm windows/386 windows/amd64"
#platforms="linux/amd64/linux/x64"
#platforms="darwin/amd64/macosx/x64"
if [ "${full_build}" == "true" ]; then
    echo "Building ${filename}:${build_version} for all platforms..."
    platforms="darwin/amd64/macosx/x64 linux/386/linux/i586 linux/amd64/linux/x64 windows/386/windows/i586 windows/amd64/windows/x64"
else
    detectPlatformSpecificBuild
    echo "Building ${filename}:${build_version} for detected ${platform} platform..."
fi

go_executable=$(which go)
if [[ -x "$go_executable" ]] ; then
    echo "Go found in \$PATH"
else
    echo "Go not found in \$PATH"
    exit 1
fi

if [ ! $skipTest ] ; then
    echo "-------------------------------------------------------"
    echo "Go TESTS"
    echo "-------------------------------------------------------"

    go test ./...

    rc=$?
    if [ $rc -ne 0 ]; then
    echo "Testing failed!"
    exit $rc
    else
        echo "Testing Successful!"
    fi
else
    echo "Skipping Go Tests..."
fi

# run the completion.go file to get the bash completion script
# To do the string replace first build the script so that we have a consistent name
go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH


for platform in ${platforms}
do
    split=(${platform//\// })
    goos=${split[0]}
    goarch=${split[1]}
    pos=${split[2]}
    parch=${split[3]}

    # ensure output file name
    output="iamctl"
    test "$output" || output="$(basename ${target} | sed 's/\.go//')"

    # add exe to windows output
    [[ "windows" == "$goos" ]] && output="$output.exe"

    echo -en "\t - $goos/$goarch..."

    iamctl_dir_name="iamctl-$build_version"
    iamctl_archive_name="$iamctl_dir_name-$pos-$parch"
    iamctl_archive_dir="${buildPath}/${iamctl_dir_name}"
    mkdir -p $iamctl_archive_dir

    cp -r "${baseDir}/docs/README.md" $iamctl_archive_dir > /dev/null 2>&1
    cp -r "${baseDir}/LICENSE" $iamctl_archive_dir > /dev/null 2>&1


    # set destination path for binary
    iamctl_bin_dir="${iamctl_archive_dir}/bin"
    mkdir -p $iamctl_bin_dir
    destination="$iamctl_bin_dir/$output"

    GOOS=$goos GOARCH=$goarch go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -o $destination $target

    pwd=`pwd`
    cd $buildPath
    if [[ "windows" == "$goos" ]]; then
        zip -r "$iamctl_archive_name.zip" $iamctl_dir_name > /dev/null 2>&1
    else
        tar czf "$iamctl_archive_name.tar.gz" $iamctl_dir_name > /dev/null 2>&1
    fi
    rm -rf $iamctl_dir_name
    cd $pwd
    echo -en $'✔ '
    echo
done

echo "Build complete!"