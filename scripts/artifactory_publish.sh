#!/bin/bash
export GOWORK=off

set -euo pipefail;

if ! [[ "$ARTIFACT_VERSION" =~ ^(v[0-9.]+(-[0-9A-Za-z_.\\-]*)?)$ ]]; then
    echo "Ref $ARTIFACT_VERSION is not a vX.X.X tag, skipping publish";
    exit 0;
fi

TAG_NAME="${BASH_REMATCH[1]}"

echo "Installing JFrog CLI..."
JFROG_VERSION="2.75.0"
JFROG_DOWNLOAD_URL="https://releases.jfrog.io/artifactory/jfrog-cli/v2/${JFROG_VERSION}/jfrog-cli-linux-amd64/jfrog"
# We need to make sure we stash the jfrog CLI outside the current directory,
# or it will be uploaded to artifactory itself in the zip file.
curl --location --fail --silent --show-error --output ~/jfrog "$JFROG_DOWNLOAD_URL"
chmod +x ~/jfrog

# Run build to prep go.mod / sum files
cd internal
go mod tidy
echo "Preparing go.mod and go.sum files for publish"
go run build.go $ARTIFACT_VERSION
cd ..

## Configure jfrog cli
~/jfrog --version
~/jfrog config add --artifactory-url https://zdrepo.jfrog.io/zdrepo --url https://zdrepo.jfrog.io --user="$ARTIFACTORY_USERNAME" --password="$ARTIFACTORY_API_KEY" zdrepo
~/jfrog go-config --global --repo-deploy go-pkg
~/jfrog config show

publish() {
  package=$1
  echo "Publishing package: $package"
  echo "Publishing $TAG_NAME to Artifactory as $ARTIFACTORY_USERNAME"
#  ~/jfrog go-publish "$TAG_NAME" --exclusions="*test.go;*.md"
  ~/jfrog go-publish "$TAG_NAME"
  echo "Publish succeeded!"
  # sleep to guarantee that the publish save is complete before it possibly being necessary fro the next build
  sleep 5
}

# Must publish packages in order based on dependencies.
# 1. test
# 2. serialize
# 3. datastructures
# 4. ratelimit
# 5. encryption
# 6. cache
# 7. functions

cd ./test
publish test

cd ../serialize
publish serialize

cd ../datastructures
publish datastructures

cd ../ratelimit
publish ratelimit

cd ../encryption
publish encryption

cd ../cache
publish cache

cd ../functions
publish functions