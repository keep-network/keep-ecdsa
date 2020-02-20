#!/bin/bash
set -ex

# Fetch addresses of contacts migrated from keep-network projects.
# CONTRACT_DATA_BUCKET is configured in Circle CI Context config to values specific
# to the given environment.

TBTC_SYSTEM_CONTRACT_DATA="TBTCSystem.json"

function fetch_tbtc_system_contract_artifact() {
    gsutil -q cp gs://${CONTRACT_DATA_BUCKET}/tbtc/${TBTC_SYSTEM_CONTRACT_DATA} .
}

fetch_tbtc_system_contract_artifact
