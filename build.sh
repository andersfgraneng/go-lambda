#!/usr/bin/env bash
BUILD_DIR="build"

if [ ! -d "$BUILD_DIR" ]
then
    mkdir "$BUILD_DIR"
fi

GOOS=linux GOARCH=amd64 go build -o build/main
cp watermark.png "$BUILD_DIR/watermark.png"
zip "$BUILD_DIR/myfunction.zip" "$BUILD_DIR/main" "$BUILD_DIR/watermark.png"