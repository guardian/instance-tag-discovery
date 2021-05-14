#!/bin/bash
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

build() {
  # setup
  PKG_ROOT="${SCRIPT_DIR}/../target/instance-tag-discovery-${GOARCH}_1.0-1"
  mkdir -p ${PKG_ROOT}/usr/bin ${PKG_ROOT}/DEBIAN ${PKG_ROOT}/var/lib/cloud/scripts/per-instance

  # build the binary for the given arch and os
  GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${PKG_ROOT}/usr/bin/instance-tag-discovery .

  # a script to execute this at instance boot time
  # we choose per-instance so it runs on first boot of a given instance ID
  cat > ${PKG_ROOT}/var/lib/cloud/scripts/per-instance/00-instance-tag-discovery.sh <<EOF
#!/bin/bash
set -e
/usr/bin/instance-tag-discovery
EOF
  chmod +x ${PKG_ROOT}/var/lib/cloud/scripts/per-instance/00-instance-tag-discovery.sh

  # make a control file for the package
  cat > ${PKG_ROOT}/DEBIAN/control <<EOF
Package: instance-tag-discovery
Version: 1.0-1
Section: base
Priority: optional
Architecture: ${GOARCH}
Maintainer: DevX <devx@theguardian.com>
Description: Writes out instance tags at boot time
EOF

  # build the package, if we have the right tooling
  if hash dpkg-deb 2>/dev/null; then
    dpkg-deb --build ${PKG_ROOT}
  else
    echo "Skipping packaging as dpkg-deb command doesn't exist. If you're testing this on macOS, try "
  fi
}
GOOS=linux GOARCH=amd64 build
GOOS=linux GOARCH=arm64 build
