#!/bin/bash

# Compiling gcvis
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go build -o $DIR/gcvis $DIR/../*.go

gcvis godoc -index -http=:6060
