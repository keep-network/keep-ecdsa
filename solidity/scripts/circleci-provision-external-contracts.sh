#!/bin/bash
set -ex

# Fetch contacts migrated from keep-network/keep-core project.
# TOKEN_STAKING_CONTRACT_DATA and REGISTRY_CONTRACT_DATA are set in the CircleCI
# job config to the filenames of respective contracts' artifacts.

function fetch_token_staking_contract() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${TOKEN_STAKING_CONTRACT_DATA} ./solidity/build/contracts/${TOKEN_STAKING_CONTRACT_DATA}
}

function fetch_registry_contract() {
  gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/keep-core/${REGISTRY_CONTRACT_DATA} ./solidity/build/contracts/${REGISTRY_CONTRACT_DATA}
}

fetch_token_staking_contract
fetch_registry_contract
