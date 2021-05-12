#!/bin/sh
set -e
mkdir target
build() {
  PKG_ROOT="target/instance-tag-discovery-${LINUXARCH}_1.0-1"
  mkdir -p ${PKG_ROOT}/usr/bin ${PKG_ROOT}/DEBIAN
  go build -o ${PKG_ROOT}/usr/bin/instance-tag-discovery .
  cat > ${PKG_ROOT}/DEBIAN/control <<EOF
Package: instance-tag-discovery
Version: 1.0-1
Section: base
Priority: optional
Architecture: ${LINUXARCH}
Maintainer: DevX <devx@theguardian.com>
Description: Writes out instance tags at boot time
EOF
  if hash dpkg-deb 2>/dev/null; then
    dpkg-deb -build ${PKG_ROOT}
  else
    echo "Skipping packaging as dpkg-deb command doesn't exist"
  fi
}
GOOS=linux GOARCH=amd64 LINUXARCH=x86_64 build
GOOS=linux GOARCH=arm64 LINUXARCH=aarch64 build
