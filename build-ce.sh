#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
OD="$(pwd)"

# Pushes application version into the build information.
RR_VERSION=1.4.2

# Hardcode some values to the core package
LDFLAGS="$LDFLAGS -X github.com/spiral/roadrunner/cmd/rr/cmd.Version=${RR_VERSION}"
LDFLAGS="$LDFLAGS -X github.com/spiral/roadrunner/cmd/rr/cmd.BuildTime=$(date +%FT%T%z)"

build(){
	echo Packaging $1 CE Build
	bdir=roadrunner-ce-${RR_VERSION}-$2-$3
	rm -rf builds/$bdir && mkdir -p builds/$bdir
	GOOS=$2 GOARCH=$3 ./build-ce.sh

	if [ "$2" == "windows" ]; then
		mv rr-ce builds/$bdir/rr-ce.exe
	else
		mv rr-ce builds/$bdir
	fi

	cp README.md builds/$bdir
	cp CHANGELOG.md builds/$bdir
	cp LICENSE builds/$bdir
	cd builds

	if [ "$2" == "linux" ]; then
		tar -zcf $bdir.tar.gz $bdir
	else
		zip -r -q $bdir.zip $bdir
	fi

	rm -rf $bdir
	cd ..
}

if [ "$1" == "all" ]; then
	rm -rf builds/*ce*
	build "Windows" "windows" "amd64"
	build "Mac" "darwin" "amd64"
	build "Linux" "linux" "amd64"
	build "FreeBSD" "freebsd" "amd64"
	exit
fi

CGO_ENABLED=0 go build -ldflags "$LDFLAGS -extldflags '-static'" -o "$OD/rr-ce" cmd/rr-ce/main.go
