#!/bin/bash

declare -r NEED_FMT="$(gofmt -l **/*.go)"

if [[ -z "${NEED_FMT}" ]]; then
	exit 0
fi

echo "The following files are NOT formatted with gofmt:"
echo "${NEED_FMT}"
exit 1
