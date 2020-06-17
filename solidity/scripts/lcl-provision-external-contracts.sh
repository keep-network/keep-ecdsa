#!/bin/bash
set -e

# Fetch addresses of contacts migrated from keep-network/keep-core project.
# The `keep-core` contracts have to be migrated before running this script.
# It requires `KEEP_CORE_SOL_ARTIFACTS_PATH` variable to pointing to a directory where
# contracts artifacts after migrations are located. It also expects NETWORK_ID
# variable to be set to the ID of the network where contract were deployed.
# 
# Sample command:
# KEEP_CORE_SOL_ARTIFACTS_PATH=~/go/src/github.com/keep-network/keep-core/contracts/solidity/build/contracts \
# NETWORK_ID=1801 \
#   ./lcl-provision-external-contracts.sh

REGISTRY_CONTRACT_DATA="KeepRegistry.json"
REGISTRY_PROPERTY="RegistryAddress"

TOKEN_STAKING_CONTRACT_DATA="TokenStaking.json"
TOKEN_STAKING_PROPERTY="TokenStakingAddress"

TOKEN_GRANT_CONTRACT_DATA="TokenGrant.json"
TOKEN_GRANT_PROPERTY="TokenGrantAddress"

RANDOM_BEACON_CONTRACT_DATA="KeepRandomBeaconService.json"
RANDOM_BEACON_PROPERTY="RandomBeaconAddress"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

ADDRESS_REGEXP=^0[xX][0-9a-fA-F]{40}$

# Query to get address of the deployed contract for the first network on the list.
JSON_QUERY=".networks.\"${NETWORKID}\".address"

SED_SUBSTITUTION_REGEXP="['\"][a-zA-Z0-9]*['\"]"

FAILED=false

function fetch_registry_contract_address() {
  echo "Fetching value for ${REGISTRY_PROPERTY}..."
  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$REGISTRY_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')

  if [[ !($ADDRESS =~ $ADDRESS_REGEXP) ]]; then
    echo "Invalid address: ${ADDRESS}"
    FAILED=true
  else
    echo "Found value for ${REGISTRY_PROPERTY} = ${ADDRESS}"
    sed -i -e "/${REGISTRY_PROPERTY}/s/${SED_SUBSTITUTION_REGEXP}/\"${ADDRESS}\"/" $DESTINATION_FILE
  fi
}

function fetch_token_staking_contract_address() {
  echo "Fetching value for ${TOKEN_STAKING_PROPERTY}..."

  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$TOKEN_STAKING_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')

  if [[ !($ADDRESS =~ $ADDRESS_REGEXP) ]]; then
    echo "Invalid address: ${ADDRESS}"
    FAILED=true
  else 
    echo "Found value for ${TOKEN_STAKING_PROPERTY} = ${ADDRESS}"
    sed -i -e "/${TOKEN_STAKING_PROPERTY}/s/${SED_SUBSTITUTION_REGEXP}/\"${ADDRESS}\"/" $DESTINATION_FILE
  fi
}

function fetch_token_grant_contract_address() {
  echo "Fetching value for ${TOKEN_GRANT_PROPERTY}..."

  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$TOKEN_GRANT_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')

  if [[ !($ADDRESS =~ $ADDRESS_REGEXP) ]]; then
    echo "Invalid address: ${ADDRESS}"
    FAILED=true
  else 
    echo "Found value for ${TOKEN_GRANT_PROPERTY} = ${ADDRESS}"
    sed -i -e "/${TOKEN_GRANT_PROPERTY}/s/${SED_SUBSTITUTION_REGEXP}/\"${ADDRESS}\"/" $DESTINATION_FILE
  fi
}

function fetch_random_beacon_contract_address() {
  echo "Fetching value for ${RANDOM_BEACON_PROPERTY}..."

  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$RANDOM_BEACON_CONTRACT_DATA)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')

  if [[ !($ADDRESS =~ $ADDRESS_REGEXP) ]]; then
    echo "Invalid address: ${ADDRESS}"
    FAILED=true
  else
  echo "Found value for ${TOKEN_STAKING_PROPERTY} = ${ADDRESS}"
  sed -i -e "/${RANDOM_BEACON_PROPERTY}/s/${SED_SUBSTITUTION_REGEXP}/\"${ADDRESS}\"/" $DESTINATION_FILE
  fi
}

fetch_registry_contract_address
fetch_token_staking_contract_address
fetch_token_grant_contract_address
fetch_random_beacon_contract_address

if $FAILED; then
echo "Failed to fetch external contract addresses!"
  exit 1
fi
