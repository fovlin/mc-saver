#!/bin/bash
rm -r build/mc-saver*

export GOARCH=amd64
export GOOS=linux
go build -o build/mc-saver_linux_amd64/mc-saver .

export GOOS=windows
go build -o build/mc-saver_windows_amd64/mc-saver.exe  .

export GOOS=darwin
go build -o build/mc-saver_darwin_amd64/mc-saver  .

export GOARCH=arm64
export GOOS=linux
go build -o build/mc-saver_linux_arm64/mc-saver  .

export GOOS=windows
go build -o build/mc-saver_windows_arm64/mc-saver.exe  .

export GOOS=darwin
go build -o build/mc-saver_darwin_arm64/mc-saver  .

cd build

list=$(ls)

for name in ${list[@]}; do
    tar -czvf ${name}.tar.gz ${name}
done
