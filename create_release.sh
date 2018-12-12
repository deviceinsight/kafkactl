#!/bin/bash

version=$1

if [ "$version" == "" ]; then
  echo "enter release version:"
  read version
fi

version_go="cmd/version.go"

echo "updating $version_go to version $version..."
eval "sed -i 's/const VERSION = .*/const VERSION = \"$version\"/g' $version_go"
echo "done"
