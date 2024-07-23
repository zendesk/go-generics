#!/bin/bash

set -euo pipefail;

if ! [[ "$ARTIFACT_VERSION" =~ ^(v[0-9.]+(-[0-9A-Za-z_.\\-]*)?)$ ]]; then
    echo "Ref $ARTIFACT_VERSION is not a vX.X.X tag, skipping publish";
    exit 0;
fi

TAG_NAME="${BASH_REMATCH[1]}"

echo "Installing JFrog CLI..."
JFROG_VERSION="1.53.1"
JFROG_DOWNLOAD_URL="https://releases.jfrog.io/artifactory/jfrog-cli/v1/${JFROG_VERSION}/jfrog-cli-linux-amd64/jfrog"
# We need to make sure we stash the jfrog CLI outside the current directory,
# or it will be uploaded to artifactory itself in the zip file.
curl --location --fail --silent --show-error --output ~/jfrog "$JFROG_DOWNLOAD_URL"
chmod +x ~/jfrog

echo "Publishing $TAG_NAME to Artifactory as $ARTIFACTORY_USERNAME"
export JFROG_CLI_OFFER_CONFIG=false
~/jfrog rt go-publish go-pkg "$TAG_NAME" --url=https://zdrepo.jfrog.io/zdrepo --user="$ARTIFACTORY_USERNAME" --password="$ARTIFACTORY_API_KEY"
echo "Publish succeeded!"
