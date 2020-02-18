#!/bin/bash
set -e

# Dafault inputs.
CLIENT_APP_ADDRESS_DEFAULT="0x2AA420Af8CB62888ACBD8C7fAd6B4DdcDD89BC82"
CONFIG_FILE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../configs/config.toml)

# Read user inputs.
read -p "Enter path to keep-ecdsa client config [$CONFIG_FILE_PATH_DEFAULT]: " config_file_path
CONFIG_FILE_PATH=${config_file_path:-$CONFIG_FILE_PATH_DEFAULT}

read -p "Enter client application address [$CLIENT_APP_ADDRESS_DEFAULT]: " client_app_address
CLIENT_APP_ADDRESS=${client_app_address:-$CLIENT_APP_ADDRESS_DEFAULT}

# Run script.
LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

printf "${LOG_START}Starting initialization...${LOG_END}"
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_CONFIG_FILE_PATH=$(realpath $CONFIG_FILE_PATH)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Configuring external client contract address...${LOG_END}"
CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    ./scripts/lcl-set-client-address.sh

printf "${LOG_START}Initializing contracts...${LOG_END}"
truffle exec scripts/lcl-initialize.js --network local

printf "${LOG_START}Updating keep-ecdsa client config...${LOG_END}"
KEEP_ECDSA_CONFIG_FILE_PATH=$KEEP_ECDSA_CONFIG_FILE_PATH \
    CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    truffle exec scripts/lcl-client-config.js --network local
