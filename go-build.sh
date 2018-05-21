#!/usr/bin/env bash

prg_name="server"

package=$1
if [[ -z "$package" ]]; then
  package="."
  echo "usage: $0 <package-name>"
  echo "continue with current directory"
fi
package_split=(${package//\// })
package_name=${package_split[-1]}

platforms=("linux/amd64" "linux/arm" "windows/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='./bin/'$prg_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi  

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
