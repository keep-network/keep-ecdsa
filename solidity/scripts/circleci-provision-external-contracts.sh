#!/bin/bash
set -ex

# TokenStakingAddress: Migration from keep-network/keep-core
# TOKEN_STAKING_CONTRACT_DATA is set in the CircleCI job config
# ETH_NETWORK_ID is set in the CircleCI context for each deployed environment
TOKEN_STAKING_ADDRESS=""

function fetch_token_staking_address() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${TOKEN_STAKING_CONTRACT_DATA} ./
  TOKEN_STAKING_ADDRESS=$(cat ./${TOKEN_STAKING_CONTRACT_DATA} | jq ".networks[\"${ETH_NETWORK_ID}\"].address" | tr -d '"')
}

function set_token_staking_address() {
  sed -i -e "/TokenStakingAddress/s/0x[a-fA-F0-9]\{0,40\}/${TOKEN_STAKING_ADDRESS}/" ./solidity/migrations/externals.js
}

fetch_token_staking_address
set_token_staking_address
