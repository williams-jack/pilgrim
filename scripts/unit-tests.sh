#!/bin/bash

status=0

for pkg in $(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...); do
  go test -v "$pkg"
  if [ $? -ne 0 ]; then
    status=1
  fi
done

exit $status
