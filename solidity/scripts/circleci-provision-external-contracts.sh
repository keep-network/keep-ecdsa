#!/bin/bash
set -ex

# Fetch addresses of contacts migrated from keep-network/keep-core project.
# CONTRACT_DATA_BUCKET and ETH_NETWORK_ID are configured in Circle CI Context 
# config to values specific to given environment.

REGISTRY_CONTRACT_DATA="KeepRegistry.json"
REGISTRY_PROPERTY="RegistryAddress"

TOKEN_STAKING_CONTRACT_DATA="TokenStaking.json"
TOKEN_STAKING_PROPERTY="TokenStakingAddress"

RANDOM_BEACON_CONTRACT_DATA="KeepRandomBeaconService.json"
RANDOM_BEACON_PROPERTY="RandomBeaconAddress"

# Query to get address of the deployed contract for the specific network.
JSON_QUERY=".networks[\"${ETH_NETWORK_ID}\"].address"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

function fetch_registry_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${REGISTRY_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${REGISTRY_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${REGISTRY_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_token_staking_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${TOKEN_STAKING_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${TOKEN_STAKING_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${TOKEN_STAKING_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

function fetch_random_beacon_contract_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${RANDOM_BEACON_CONTRACT_DATA} .
  local ADDRESS=$(cat ./${RANDOM_BEACON_CONTRACT_DATA} | jq "$JSON_QUERY" | tr -d '"')
  sed -i -e "/${RANDOM_BEACON_PROPERTY}/s/0x[a-fA-F0-9]\{0,40\}/${ADDRESS}/" $DESTINATION_FILE
}

fetch_registry_contract_address
fetch_token_staking_contract_address
fetch_random_beacon_contract_address
