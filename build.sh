#!/bin/sh
#$(aws ecr get-login --no-include-email --region us-west-2)

go mod tidy 
gofmt -s -w src/*/*.go
go vet src/genericjson/*.go
golint src/*/*.go
echo "\nrequire github.com/ravenbill/genericjson v0.0.0" >> go.mod
go test -v github.com/ravenbill/genericjson
if [ $? -ne 0 ]; then
    exit
fi
