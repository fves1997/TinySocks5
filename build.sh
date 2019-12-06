#!/usr/bin/env bash
export GOPATH=${PWD}
local="cmd/local.go"
server="cmd/server.go"

if [[ ${1} == "linux" ]]; then
    export GOOS=linux
    export GOARCH=amd64
    echo "linux"
    go build -v -ldflags "-s -w" -o bin/server-linux ${local}
    go build -v -ldflags "-s -w" -o bin/server-linux ${server}
elif [[ ${1} == "android" ]];then
    export GOOS=android
    export GOARCH=arm
    echo "android"
    go build -v -ldflags "-s -w" -o bin/local-android ${local}
    go build -v -ldflags "-s -w" -o bin/server-android ${server}
elif [[ ${1} == "windows" ]];then
    export GOOS=windows
    export GOARCH=amd64
    echo "windows"
    go build -v -ldflags "-s -w" -o bin/local-windows.exe ${local}
    go build -v -ldflags "-s -w" -o bin/server-windows.exe ${server}
else
    echo "error"
fi
