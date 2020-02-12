#!/bin/bash
# This script copies contracts artifacts from provided local storage destination.
# It expects an `KEEP_CORE_ARTIFACTS` variable to be set to a path to the source
# directory holding the artifacts.
#
# Example:
#   KEEP_CORE_ARTIFACTS=~/go/src/github.com/keep-network/keep-core/contracts/solidity/build/contracts \
#     ./solidity/scripts/lcl-copy-contracts.sh

set -e

CONTRACTS_NAMES=("Registry.json" "TokenStaking.json")

if [[ -z $KEEP_CORE_ARTIFACTS ]]; then
  echo "one or more required variables are undefined"
  exit 1
fi

SOURCE_DIR=$(realpath $KEEP_CORE_ARTIFACTS)
DESTINATION_DIR=$(realpath $(dirname $0)/../build/contracts/)

function create_destination_dir() {
  mkdir -p $DESTINATION_DIR
}

function copy_contracts() {
  for CONTRACT_NAME in ${CONTRACTS_NAMES[@]}
  do
    cp $(realpath $SOURCE_DIR/$CONTRACT_NAME) $DESTINATION_DIR/
    echo "Copied contract: [${CONTRACT_NAME}]"
  done
}

echo "Copy contracts artifacts"
echo "Source: [${SOURCE_DIR}]"
echo "Destination: [${DESTINATION_DIR}]"
create_destination_dir
copy_contracts
