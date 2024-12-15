#!/bin/bash

# Generate src/version.go with the version number provided as an argument

# Check if gomplate is installed
if ! [ -x "$(command -v gomplate)" ]; then
  echo "Error: gomplate is not installed." >&2
  exit 1
fi

# Check if the version is provided
if [ -z "$1" ]; then
  echo "Please provide a version number"
  exit 1
fi

# Echo the version number without trailing 
printf "Generate src/version.go with version: $1\n"
# Echo the version number without trailing and pipe it to gomplate to generate the version.go file
printf "$1" | gomplate -f scripts/.templates/version.tpl.go -d version=stdin: > src/version.go
