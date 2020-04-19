#!/bin/bash
set -e

PGM="gofwd"
DUOINI="duo.ini"
DUOUSER="ChangeMe"
CHROOT="${HOME}/gofwd_chroot"
FROM="1.2.3.4:4567"
TO="192.168.1.1:22"

if [ ! -e "${DUOINI}" ] ; then
    echo "File not found: ${DUOINI}"
    echo "Exiting."
fi

# statically compile gofwd, linux command:
go build -tags netgo -ldflags '-extldflags "-static"'
echo "Checking for static build:"
set +e
ldd ${PGM}
set -e
echo

if [ ! -e "${PGM}" ] ; then
    echo "File not found: ${PGM}"
    echo "Exiting."
fi

if [ ! -e ${CHROOT}/etc/pki/ca-trust/extracted ] ; then
    mkdir -p ${CHROOT}/etc/pki/ca-trust/extracted
    cp ${PGM} ${CHROOT}
    cp ${DUOINI} ${CHROOT}
    cp -a /etc/resolv.conf ${CHROOT}/etc/
    cp -a /etc/pki/ca-trust/extracted/ ${CHROOT}/etc/pki/ca-trust/
fi

# Add the -p switch, if you are running in a NAT environment such as AWS
sudo chroot ${CHROOT} /${PGM} -f ${FROM} -t ${TO} --duo /${DUOINI}:${DUOUSER}

