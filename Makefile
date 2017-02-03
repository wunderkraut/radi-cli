.PHONY: cli-build build install clean

MAKE_SCRIPT="./build/make.sh"

default: all

all: clean getdeps fmt build install

build: clean getdeps fmt build install

local: fmt binary install

fmt:
	${MAKE_SCRIPT} fmt

build:
	${MAKE_SCRIPT} binary

updatedeps:
	${MAKE_SCRIPT} update-deps

install:
	${MAKE_SCRIPT} install

clean:
	${MAKE_SCRIPT} clean
