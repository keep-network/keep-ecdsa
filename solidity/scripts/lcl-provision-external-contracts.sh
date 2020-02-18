#!/bin/bash
set -ex

# Fetch addresses of contacts migrated from keep-network/keep-core project.
# The `keep-core` contracts have to be migrated before running this script.
# It requires `KEEP_CORE_SOL_ARTIFACTS_PATH` variable to pointing to a directory where
# contracts artifacts after migrations are located.
# It takes address of the first network entry in the migration artifact.
# 
# Sample command:
#   KEEP_CORE_SOL_ARTIFACTS_PATH=~/go/src/github.com/keep-network/keep-core/contracts/solidity/build/contracts \
#   ./lcl-provision-external-contracts.sh

REGISTRY_CONTRACT_DATA="Registry.json"
REGISTRY_PROPERTY="RegistryAddress"

TOKEN_STAKING_CONTRACT_DATA="TokenStaking.json"
TOKEN_STAKING_PROPERTY="TokenStakingAddress"

RANDOM_BEACON_CONTRACT_DATA="KeepRandomBeaconService.json"
RANDOM_BEACON_PROPERTY="RandomBeaconAddress"

# Query to get address of the deployed contract for the first network on the list.
JSON_QUERY="[.networks[].address][0]"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

function fetch_registry_contract_address() {
  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$REGISTRY_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')
  sed -i -e "/${REGISTRY_PROPERTY}/s/'[a-zA-Z0-9]*'/'${ADDRESS}'/" $DESTINATION_FILE
}

function fetch_token_staking_contract_address() {
  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$TOKEN_STAKING_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')
  sed -i -e "/${TOKEN_STAKING_PROPERTY}/s/'[a-zA-Z0-9]*'/'${ADDRESS}'/" $DESTINATION_FILE
}

function fetch_random_beacon_contract_address() {
  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$RANDOM_BEACON_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')
  sed -i -e "/${RANDOM_BEACON_PROPERTY}/s/'[a-zA-Z0-9]*'/'${ADDRESS}'/" $DESTINATION_FILE
}

fetch_registry_contract_address
fetch_token_staking_contract_address
fetch_random_beacon_contract_address
