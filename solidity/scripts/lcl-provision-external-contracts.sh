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

KEEP_TOKEN_CONTRACT_DATA="KeepToken.json"
KEEP_TOKEN_PROPERTY="KeepTokenAddress"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

ADDRESS_REGEXP=^0[xX][0-9a-fA-F]{40}$

# Query to get address of the deployed contract for the first network on the list.
JSON_QUERY=".networks.\"${NETWORKID}\".address"

SED_SUBSTITUTION_REGEXP="['\"][a-zA-Z0-9]*['\"]"

FAILED=false

function fetch_contract_address() {
  property_name=$1
  contract_data=$2

  echo "Fetching value for ${property_name}..."

  local contractDataPath=$(realpath $KEEP_CORE_SOL_ARTIFACTS_PATH/$contract_data)
  local ADDRESS=$(cat ${contractDataPath} | jq "${JSON_QUERY}" | tr -d '"')

  if [[ !($ADDRESS =~ $ADDRESS_REGEXP) ]]; then
    echo "Invalid address: ${ADDRESS}"
    FAILED=true
  else 
    echo "Found value for ${property_name} = ${ADDRESS}"
    sed -i -e "/${property_name}/s/${SED_SUBSTITUTION_REGEXP}/\"${ADDRESS}\"/" $DESTINATION_FILE
  fi
}

fetch_contract_address $REGISTRY_PROPERTY $REGISTRY_CONTRACT_DATA
fetch_contract_address $TOKEN_STAKING_PROPERTY $TOKEN_STAKING_CONTRACT_DATA
fetch_contract_address $TOKEN_GRANT_PROPERTY $TOKEN_GRANT_CONTRACT_DATA
fetch_contract_address $RANDOM_BEACON_PROPERTY $RANDOM_BEACON_CONTRACT_DATA
fetch_contract_address $KEEP_TOKEN_PROPERTY $KEEP_TOKEN_CONTRACT_DATA

if $FAILED; then
echo "Failed to fetch external contract addresses!"
  exit 1
fi
