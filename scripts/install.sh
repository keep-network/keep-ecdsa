#!/bin/bash
set -e pipefail

GOPRIVATE="keep-network/keep-core"

# Dafault inputs.
KEEP_ETHEREUM_PASSWORD_DEFAULT="password"
KEEP_CORE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../../keep-core)

# Read user inputs.
read -p "Enter ethereum accounts password [$KEEP_ETHEREUM_PASSWORD_DEFAULT]: " ethereum_password
KEEP_ETHEREUM_PASSWORD=${ethereum_password:-$KEEP_ETHEREUM_PASSWORD_DEFAULT}

read -p "Enter path to the keep-core project [$KEEP_CORE_PATH_DEFAULT]: " keep_core_path
KEEP_CORE_PATH=$(realpath ${keep_core_path:-$KEEP_CORE_PATH_DEFAULT})

# Run script.
LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

printf "${LOG_START}Starting installation...${LOG_END}"
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)
KEEP_CORE_SOL_PATH=$(realpath $KEEP_CORE_PATH/solidity)
KEEP_CORE_SOL_ARTIFACTS_PATH=$(realpath $KEEP_CORE_SOL_PATH/build/contracts)

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install

printf "${LOG_START}Unlocking ethereum accounts...${LOG_END}"
KEEP_ETHEREUM_PASSWORD=$KEEP_ETHEREUM_PASSWORD \
    npx truffle exec scripts/unlock-eth-accounts.js --network local

printf "${LOG_START}Finding current ethereum network ID...${LOG_END}"

output=$(npx truffle exec ./scripts/get-network-id.js --network local)
NETWORKID=$(echo "$output" | tail -1)
printf "Current network ID: ${NETWORKID}\n"

printf "${LOG_START}Fetching external contracts addresses...${LOG_END}"
KEEP_CORE_SOL_ARTIFACTS_PATH=$KEEP_CORE_SOL_ARTIFACTS_PATH \
NETWORKID=$NETWORKID \
    ./scripts/lcl-provision-external-contracts.sh

printf "${LOG_START}Migrating contracts...${LOG_END}"
npm run clean
npx truffle migrate --reset --network local

printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
cd $KEEP_ECDSA_PATH
go generate ./...
go build -a -o keep-ecdsa .

printf "${DONE_START}Installation completed!${DONE_END}"
