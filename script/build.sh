#!/bin/sh
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

mkdir ${SCRIPT_DIR}/target
build() {
  PKG_ROOT="${SCRIPT_DIR}/target/instance-tag-discovery-${GOARCH}_1.0-1"
  mkdir -p ${PKG_ROOT}/usr/bin ${PKG_ROOT}/DEBIAN
  GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${PKG_ROOT}/usr/bin/instance-tag-discovery .
  cat > ${PKG_ROOT}/DEBIAN/control <<EOF
Package: instance-tag-discovery
Version: 1.0-1
Section: base
Priority: optional
Architecture: ${GOARCH}
Maintainer: DevX <devx@theguardian.com>
Description: Writes out instance tags at boot time
EOF
  cp ${SCRIPT_DIR}/instance-tag-discovery.service ${PKG_ROOT}/etc/systemd/system/instance-tag-discovery.service
  if hash dpkg-deb 2>/dev/null; then
    dpkg-deb --build ${PKG_ROOT}
  else
    echo "Skipping packaging as dpkg-deb command doesn't exist"
  fi
}
GOOS=linux GOARCH=amd64 build
GOOS=linux GOARCH=arm64 build
