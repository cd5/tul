#!/bin/bash
set -e
D="$(dirname $(realpath $0))"
cd "$D"
mkdir -p "bin"
echo "Build $(realpath --relative-base=$PWD $D/bin/tulgo)"
go build -o "$D/bin/tulgo" *.go
