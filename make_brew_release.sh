#!/usr/bin/env bash
# expects the following variables
# $1 BINARY_NAME
# $2 VERSION


# See https://gist.github.com/maelvalais/068af21911c7debc4655cdaa41bbf092 for a rough guide on CI/CD for Brew.
# This script runs the following process:
# 1. Tap the repo in case it's not yet tapped.
# 2. Disregard all changes and switch to master.
# 3. Create a new formula from the template with the correct version.
# 4. Get the correct SHA256 sum for the version and update the formula.
# 5. Build the bottle for the current operating system.
# 6. Create a final formula with the correct version, SHA256 and bottle info.
# 7. Commit to a new branch and push.

# setup error handling
set -e -o pipefail

# set some variables used below
BINARY_NAME=$1
VERSION=$2
SOURCE_DIR=$(pwd)


function cleanup() {
  set -x
  rm -f "${BINARY_NAME}".rb.bottle* "${BINARY_NAME}".rb.bak
}

trap cleanup EXIT

# create a signed zip file
# make sign

SHA256_ZIP_DARWIN=$(sha256sum "assets/${BINARY_NAME}-darwin-amd64.zip" | awk '{ print $1 }' )

# add tap in case it's missing
brew tap allcloud-io/tools
TAP_DIR=$(brew --repo allcloud-io/tools)
# change to tap directory
cd "$TAP_DIR" || exit 1
# stash all changes so we have a clean working directory
git clean -d -x -f
git reset --hard
git fetch --all
git checkout master
git pull

# set the correct version
sed "s:%VERSION%:${VERSION}:" "${BINARY_NAME}.rb.template" | sed "s:%BOTTLE%::" > "${BINARY_NAME}.rb"
# and calc sha256
SHA256=$(brew fetch "${BINARY_NAME}" --build-from-source 2>/dev/null | grep SHA256 | cut -d" " -f2 || true)

# replace version and sha256 placeholder in template
sed "s:%VERSION%:${VERSION}:" "${BINARY_NAME}.rb.template" | \
sed "s:%SOURCE_SHA%:${SHA256}:" | \
sed "s:%BUILD_DARWIN_SHA%:${SHA256_ZIP_DARWIN}:" > "${BINARY_NAME}.rb"

# change back to original workdir
cd "$SOURCE_DIR" || exit 1
# build the bottle
brew test-bot "allcloud-io/tools/${BINARY_NAME}"

# commit to git and push to origin
BRANCHNAME=auto/${BINARY_NAME}-${VERSION}
git checkout -b "$BRANCHNAME" || git checkout "$BRANCHNAME"
git add "${BINARY_NAME}.rb"
git commit -m "Automatic commit of bottle build for version $VERSION of $BINARY_NAME."
git push origin "$BRANCHNAME"
