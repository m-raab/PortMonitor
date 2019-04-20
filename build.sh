#!/usr/bin/env bash

package=PortMonitor.go
package_name=portMonitor

platforms=("linux/amd64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=bin/${GOOS}-${GOARCH}/${package_name}

    env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${output_name} ${package}
    if [ $? -ne 0 ]; then
        echo 'An error has occurred during GO compilation! Aborting the script execution...'
        exit 1
    fi
    mkdir -p dist && tar -zcvf dist/${package_name}.${GOOS}.${GOARCH}.tar.gz bin/${GOOS}-${GOARCH}/${package_name}
    if [ $? -ne 0 ]; then
        echo 'An error has occurred during packaging! Aborting the script execution...'
        exit 1
    fi
done