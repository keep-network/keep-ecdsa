#!/bin/bash
set -ex

# Migration from keep-network/keep-core
# TOKEN_STAKING_CONTRACT_DATA is set in the CircleCI job config
# ETH_NETWORK_ID is set in the CircleCI context for each deployed environment

function fetch_token_staking_contract() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${TOKEN_STAKING_CONTRACT_DATA} ../build/contracts/
}

function fetch_registry_contract() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${REGISTRY_CONTRACT_DATA} ../build/contracts/
}

fetch_token_staking_contract
fetch_registry_contract
