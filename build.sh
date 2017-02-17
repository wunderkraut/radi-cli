#!/bin/bash

#
# Build wundertools in a container
#
# @NOTE to specify a different os/arch:
#    - GOOS : linux darwin windows
#    - GOARCH : amd64 arm arm64
#
# @NOTE !does not install it yet
#  (installs it, but inside the container)
#

source build/.os-detect
source build/make.sh

EXEC_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INTERNAL_LIBRARY_PATH="github.com/wunderkraut/radi-cli"

echo "***** Building RADI cli client.

This will build the radi-cli as a 'radi' binary for '$GOOS-$GOARCH'. 

(Override this by setting \$GOOS and \$GOARCH environment variables)

 **** Building in containerized golang environment
 "

# some sanity stuff, to prevent docker related permissions issues
mkdir -p "${RADI_BUILD_PATH}"
mkdir -p ".git/modules/vendor"
chmod u+x Makefile

# Run the build inside a container
#
#  - volumify the submodule changes
#  - build in a valid gopath to get active vendor dependencies
#  - pass in env variables for environment control
docker run --rm -ti \
	-v "${EXEC_PATH}:/go/src/${INTERNAL_LIBRARY_PATH}:Z" \
	-v "/go/src/${INTERNAL_LIBRARY_PATH}/.git/modules/vendor" \
	-v "/go/src/${INTERNAL_LIBRARY_PATH}/vendor" \
	-e "GOOS=${GOOS}" \
	-w "/go/src/${INTERNAL_LIBRARY_PATH}" \
	golang:${GOVERSION} \
	make build

echo " 

Exited container
"

echo " **** Containerized build complete 

an executable binary has (hopefully) now been built 
in ${RADI_BUILD_BINARY_PATH}

"

# @TODO implement some improved logic for determining
#    Install path, and sudo

export RADI_INSTALL_PATH="/usr/local/bin"

echo " **** Installation

This installer can now install the built binary for you,
if you don't want to do it manually.

The planned installation path is : ${RADI_INSTALL_PATH}

Would you like to me install a binary to that location? (y/n)
"
read  yninstall
case "$yninstall" in
    [Yy]* )

		if [ -w "RADI_INSTALL_PATH" ] ; then 
			export RADI_INSTALL_SUDO=""
		else 
			export RADI_INSTALL_SUDO="`which sudo`  -E"
			echo "--> detected that sudo will be required, as you don't have write privelege to the target path"
		fi

		${RADI_INSTALL_SUDO} make install

		;;
    *)
		echo " "
		echo "skipped installation"
		;;
esac
