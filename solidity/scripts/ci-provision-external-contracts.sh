#!/bin/bash
set -ex

# Fetch addresses of contacts migrated from keep-network/keep-core project.
# CONTRACT_DATA_BUCKET and ETH_NETWORK_ID should be passed as environment
# variables straight from the CI context.

if [[ -z $CONTRACT_DATA_BUCKET || -z $ETH_NETWORK_ID ]]; then
  echo "one or more required variables are undefined"
  exit 1
fi

if ! [ -x "$(command -v jq)" ]; then echo "jq is not installed"; exit 1; fi

: ${CONTRACT_DATA_BUCKET_DIR:=keep-core}

# Query to get address of the deployed contract for the specific network.
JSON_QUERY=".networks[\"${ETH_NETWORK_ID}\"].address"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

function fetch_contract_address() {
  local ARTIFACT_FILENAME=$1
  local PROPERTY_NAME=$2

  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${ARTIFACT_FILENAME} .

  local ADDRESS=$(cat ./${ARTIFACT_FILENAME} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${PROPERTY_NAME}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

fetch_contract_address "KeepRegistry.json" "RegistryAddress"
fetch_contract_address "TokenStaking.json" "TokenStakingAddress"
fetch_contract_address "TokenGrant.json" "TokenGrantAddress"
fetch_contract_address "KeepRandomBeaconService.json" "RandomBeaconAddress"
fetch_contract_address "KeepToken.json" "KeepTokenAddress"

echo "result content of $DESTINATION_FILE"
cat $DESTINATION_FILE