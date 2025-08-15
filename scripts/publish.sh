#!/bin/bash
export GOWORK=off

set -euo pipefail;

if ! [[ "$ARTIFACT_VERSION" =~ ^(v[0-9.]+(-[0-9A-Za-z_.\\-]*)?)$ ]]; then
    echo "Ref $ARTIFACT_VERSION is not a vX.X.X tag, skipping publish";
    exit 0;
fi

TAG_NAME="${BASH_REMATCH[1]}"

# Run build to prep go.mod / sum files
cd internal
go mod tidy
echo "Preparing go.mod and go.sum files for publish"
go run build.go $ARTIFACT_VERSION
cd ..
