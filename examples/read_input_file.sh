#!/bin/bash

# Compiling gcvis
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go build -o $DIR/gcvis $DIR/../*.go

cat $DIR/go15_stderr.log | $DIR/gcvis
