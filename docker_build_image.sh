#!/bin/bash

set -euo pipefail

function build() {
    echo
    echo "Building binary: ${GF}"
    echo
    if [[ -e ${GF} ]] ; then
        echo
        echo "Using existing binary: ${GF}"
        echo
        ls -l ${GF}
        return
    fi
    CGO_ENABLED=0 go build -ldflags="-s -w"
    if [[ ! -e ${GF} ]] ; then
        echo
        echo "Unable to build file: ${GF}"
        echo "Build aborted"
        echo
        exit 1
    else
        echo
        ls -l ${GF}
    fi
}

function bundle() {
    echo
    echo "Generating file: ${SC}/${BUND}"
    echo

    cd ${SC}
    curl -LOs https://raw.githubusercontent.com/curl/curl/master/scripts/mk-ca-bundle.pl
    perl mk-ca-bundle.pl
    if [[ ! -e ${BUND} ]] ; then
        echo
        echo "Unable to create file: ${BUND}"
        echo "Build aborted"
        echo
     exit 1
    else
        rm -f certdata.txt mk-ca-bundle.pl
        chmod 644 ${BUND}
        echo
        ls -l ${BUND}
        cd -
    fi
}

function image() {
    echo
    echo Creating Docker Image: ${IMG}
    echo
    sleep 1
    docker build -t ${IMG} -f Dockerfile .
    docker image ls ${IMG}
    echo
    echo
    echo Now, use ${IMG} within the docker_start_gofwd.sh script,
    echo which will need editing for your deployment.
    echo
}

if [ $# -eq 0 ] ; then
    echo give Version Tag on cmd line
    echo example: v052.1
    exit 1
fi

TAG=$1
IMG=gofwd:${TAG}
BUND="ca-bundle.crt"
SC="ssl/certs/"
GF="gofwd"
#GOOS="linux"
#GOARCH="amd64"
if [[ ! -e ${SC} ]] ; then
    mkdir -p -m 755 ${SC}
fi

build
#bundle
#image

