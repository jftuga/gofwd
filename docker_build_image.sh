#!/bin/bash

if [ $# -eq 0 ] ; then
    echo give Version Tag on cmd line
    echo example: v052.1
    exit 1
fi

TAG=$1
IMG=gofwd:${TAG}
echo
echo Creating Docker Image: ${IMG}
echo
docker build -t ${IMG} -f Dockerfile .
echo
echo Now, use ${IMG} with in the docker_start_gofwd.sh script,
echo which will need editing.
echo
