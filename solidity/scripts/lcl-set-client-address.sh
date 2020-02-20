#!/bin/bash

# Sets property of external system address to a configured `CLIENT_APP_ADDRESS`
# variable value.
# 
# Sample command:
#   CLIENT_APP_ADDRESS="0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82" \
#   ./lcl-set-client-address.sh

set -ex

TBTC_SYSTEM_PROPERTY="TBTCSystemAddress"

DESTINATION_FILE=$(realpath $(dirname $0)/../migrations/external-contracts.js)

function update_tbtc_system_address() {
  sed -i -e "/${TBTC_SYSTEM_PROPERTY}/s/'[a-zA-Z0-9]*'/'${CLIENT_APP_ADDRESS}'/" $DESTINATION_FILE
}

update_tbtc_system_address
