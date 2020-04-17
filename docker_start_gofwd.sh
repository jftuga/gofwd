#!/bin/bash

# Start gofwd in a docker container

# You will first neet to create a Docker Image with the "docker_build_image.sh" script
# Then, set the IMG variable

# This script can be run at reboot by adding the following into your crontab:
# (to edit, run: crontab -e)
# @reboot $HOME/bin/docker_start_gofwd.sh

# After launch, you can view logs:
# 1) Get the container ID: docker container ps
# 2) View the logs: docker logs -f <ID>

IMG=ChangeMe
LOG=${HOME}/.gofwd.log
DUOINI=${HOME}/duo.ini
DUOUSR=ChangeMe
EXTERNPORT=4567
FROM=1.2.3.4:${EXTERNPORT}
TO=192.168.1.1:22
LOCATION=39.858706,-104.670732
DIST=80
RESTART=on-failure:10

#echo "Starting ${IMG} at `date`"
echo "" >> ${LOG} 2>&1
echo "========================================================" >> ${LOG} 2>&1
echo "Starting ${IMG} as `date`" >> ${LOG} 2>&1
echo "========================================================" >> ${LOG} 2>&1

docker run -d --restart=${RESTART} \
    -p ${EXTERNPORT}:${EXTERNPORT} -v ${DUOINI}:/duo.ini \
    ${IMG} -f ${FROM} -t ${TO} -l ${LOCATION} -d ${DIST} -p --duo /duo.ini:${DUOUSR}

docker container ps >> ${LOG} 2>&1

