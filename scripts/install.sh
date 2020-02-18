#!/bin/bash
set -e

# Dafault inputs.
KEEP_ETHEREUM_PASSWORD_DEFAULT="password"
KEEP_CORE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../../keep-core)
CLIENT_APP_ADDRESS_DEFAULT="0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82"
CONFIG_FILE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../configs/config.toml)

# Read user inputs.
read -p "Enter ethereum accounts password [$KEEP_ETHEREUM_PASSWORD_DEFAULT]: " ethereum_password
KEEP_ETHEREUM_PASSWORD=${ethereum_password:-$KEEP_ETHEREUM_PASSWORD_DEFAULT}

read -p "Enter path to the keep-core project [$KEEP_CORE_PATH_DEFAULT]: " keep_core_path
KEEP_CORE_PATH=$(realpath ${keep_core_path:-$KEEP_CORE_PATH_DEFAULT})

read -p "Enter path to keep-ecdsa client config [$CONFIG_FILE_PATH_DEFAULT]: " config_file_path
CONFIG_FILE_PATH=${config_file_path:-$CONFIG_FILE_PATH_DEFAULT}

read -p "Enter client application address [$CLIENT_APP_ADDRESS_DEFAULT]: " client_app_address
CLIENT_APP_ADDRESS=${client_app_address:-$CLIENT_APP_ADDRESS_DEFAULT}

# Run script.
LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

printf "${LOG_START}Starting installation...${LOG_END}"
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_CONFIG_FILE_PATH=$(realpath $CONFIG_FILE_PATH)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)
KEEP_CORE_SOL_PATH=$(realpath $KEEP_CORE_PATH/contracts/solidity)
KEEP_CORE_SOL_ARTIFACTS_PATH=$(realpath $KEEP_CORE_SOL_PATH/build/contracts)

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install

printf "${LOG_START}Unlocking ethereum accounts...${LOG_END}"
KEEP_ETHEREUM_PASSWORD=$KEEP_ETHEREUM_PASSWORD \
    truffle exec scripts/unlock-eth-accounts.js --network local

printf "${LOG_START}Fetching external contracts addresses...${LOG_END}"
KEEP_CORE_SOL_ARTIFACTS_PATH=$KEEP_CORE_SOL_ARTIFACTS_PATH \
    ./scripts/lcl-provision-external-contracts.sh

CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    ./scripts/lcl-set-client-address.sh

printf "${LOG_START}Migrating contracts...${LOG_END}"
npm run clean
truffle migrate --reset --network local

printf "${LOG_START}Initializing contracts...${LOG_END}"
truffle exec scripts/lcl-initialize.js --network local

printf "${LOG_START}Updating keep-ecdsa client config...${LOG_END}"
KEEP_ECDSA_CONFIG_FILE_PATH=$KEEP_ECDSA_CONFIG_FILE_PATH \
    CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    KEEP_DATA_DIR=$KEEP_DATA_DIR \
    truffle exec scripts/lcl-client-config.js --network local

printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
cd $KEEP_ECDSA_PATH
go generate ./...
go build -a -o keep-ecdsa .
