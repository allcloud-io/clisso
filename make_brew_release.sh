#!/usr/bin/env bash
# expects the following variables
# $1 BINARY_NAME
# $2 VERSION

set -xe -o pipefail

BINARY_NAME=$1
VERSION=$2
SOURCE_DIR=$(pwd)


function cleanup() {
  rm -f ${BINARY_NAME}.rb.bottle*
}

trap cleanup EXIT

# add tap in case it's missing
brew tap allcloud-io/tools
TAP_DIR=$(brew --repo allcloud-io/tools)
# change to tap directory
cd $TAP_DIR
# stash all changes so we have a clean working directory
git stash
# calc sha256
SHA256=$(brew fetch clisso --build-from-source 2>/dev/null | grep SHA256 | cut -d" " -f2)
# replace version and sha256 placeholder in template
sed "s:%VERSION%:${VERSION}:" ${BINARY_NAME}.rb.template | \
sed "s:%SOURCE_SHA%:${SHA256}:" > ${BINARY_NAME}.rb.bottle

# generate parts to be assembled later
grep -B100 '%BOTTLE%' ${BINARY_NAME}.rb.bottle | grep -v '%BOTTLE%' > ${BINARY_NAME}.rb.bottle.head
grep -A100 '%BOTTLE%' ${BINARY_NAME}.rb.bottle | grep -v '%BOTTLE%' > ${BINARY_NAME}.rb.bottle.tail

# skip the bottle placeholder for now
cat ${BINARY_NAME}.rb.bottle.head ${BINARY_NAME}.rb.bottle.tail > ${BINARY_NAME}.rb

# change back to original workdir
cd $SOURCE_DIR
# build the bottle
brew test-bot allcloud-io/tools/${BINARY_NAME}

# create a tempfile
TEMPFILE=$(mktemp)

for json in `ls -1 *bottle.json`; do
  # extract the mac version the bottle was build for
  MAC_VERSION=$(echo $json | cut -d. -f4);
  # extract the sha256 of the bottle
  SHA=$(cat $json | jq ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.sha256")
  # get the local file name
  LOCAL=$(cat $json | jq -r ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.local_filename")
  # get the remote filename
  REMOTE=$(cat $json | jq -r ".\"allcloud-io/tools/${BINARY_NAME}\".bottle.tags.$MAC_VERSION.filename")
  # rename to the correct name
  mv $LOCAL $REMOTE
  # append to tempfile
  echo "    sha256 $SHA => :$MAC_VERSION" >> ${TEMPFILE}
  rm $json
done
# add all the bottles
cat $TAP_DIR/${BINARY_NAME}.rb.bottle.head ${TEMPFILE} $TAP_DIR/${BINARY_NAME}.rb.bottle.tail > $TAP_DIR/${BINARY_NAME}.rb


git add ${BINARY_NAME}.rb && git commit -m "Automatic commit of bottle build for version $VERSION of $BINARY_NAME."
