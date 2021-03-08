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

REGISTRY_CONTRACT_DATA="KeepRegistry.json"
REGISTRY_PROPERTY="RegistryAddress"

TOKEN_STAKING_CONTRACT_DATA="TokenStaking.json"
TOKEN_STAKING_PROPERTY="TokenStakingAddress"

TOKEN_GRANT_CONTRACT_DATA="TokenGrant.json"
TOKEN_GRANT_PROPERTY="TokenGrantAddress"

RANDOM_BEACON_CONTRACT_DATA="KeepRandomBeaconService.json"
RANDOM_BEACON_PROPERTY="RandomBeaconAddress"

KEEP_TOKEN_CONTRACT_DATA="KeepToken.json"
KEEP_TOKEN_PROPERTY="KeepTokenAddress"

# Query to get address of the deployed contract for the specific network.
JSON_QUERY=".networks[\"${ETH_NETWORK_ID}\"].address"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

function fetch_registry_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${REGISTRY_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${REGISTRY_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${REGISTRY_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_token_staking_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${TOKEN_STAKING_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${TOKEN_STAKING_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${TOKEN_STAKING_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_token_grant_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${TOKEN_GRANT_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${TOKEN_GRANT_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${TOKEN_GRANT_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_random_beacon_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${RANDOM_BEACON_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${RANDOM_BEACON_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${RANDOM_BEACON_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_keep_token_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/${CONTRACT_DATA_BUCKET_DIR}/${KEEP_TOKEN_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${KEEP_TOKEN_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${KEEP_TOKEN_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

fetch_registry_contract_address
fetch_token_staking_contract_address
fetch_token_grant_contract_address
fetch_random_beacon_contract_address
fetch_keep_token_contract_address

echo "result content of $DESTINATION_FILE"
cat $DESTINATION_FILE