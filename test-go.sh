#!/usr/bin/env bash

# export GOPATH=$GOPATH:`pwd`

package=$1
if [[ -z "$package" ]]; then
  package="./..."
  echo "usage: $0 <package-name>"
  echo "continue with current directory"
fi
package_split=(${package//\// })
package_name=${package_split[-1]}

report_dir="./report"
output_name=$report_dir'/count.out'

if [ ! -d $report_dir ]; then
    mkdir $report_dir
fi

go test -covermode=count -coverprofile=$output_name $package
if [ $? -ne 0 ]; then
    echo 'An error has occurred! Aborting the script execution...'
    exit 1
fi
go tool cover -html=$output_name
