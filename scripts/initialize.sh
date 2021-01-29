#!/bin/bash
set -euo pipefail

# Dafault config files directory.
CONFIG_DIR_PATH_DEFAULT=$(realpath -m $(dirname $0)/../configs)
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)

CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=""
CLIENT_APP_ADDRESS_DEFAULT=""
NETWORK_DEFAULT="local"

help()
{
   echo ""
   echo "Usage: $0"\
        "--keep-ecdsa-config-path <path>"\
        "--application-address <address>"\
        "--private-key <private key>"\
        "--network <network>"
   echo -e "\t--keep-ecdsa-config-path: Path to the keep-ecdsa config"
   echo -e "\t--application-address: Address of application approved by the operator"
   echo -e "\t--private-key: Contract owner's account private key"
   echo -e "\t--network: Connection network for keep-core client"
   exit 1 # Exit script after printing help
}

if [ "$0" == "-help" ]; then
  help
fi

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--keep-ecdsa-config-path")    set -- "$@" "-d" ;;
    "--application-address")       set -- "$@" "-a" ;;
    "--private-key")               set -- "$@" "-k" ;;
    "--network")                   set -- "$@" "-n" ;;
    *)                             set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "d:a:k:n:" opt
do
   case "$opt" in
      d ) config_dir_path="$OPTARG" ;;
      a ) client_app_address="$OPTARG" ;;
      k ) private_key="$OPTARG" ;;
      n ) network="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

CONFIG_DIR_PATH=${config_dir_path:-$CONFIG_DIR_PATH_DEFAULT}
CLIENT_APP_ADDRESS=${client_app_address:-$CLIENT_APP_ADDRESS_DEFAULT}
KEEP_ECDSA_CONFIG_DIR_PATH=$(realpath $CONFIG_DIR_PATH)
ACCOUNT_PRIVATE_KEY=${private_key:-$CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY}
NETWORK=${network:-$NETWORK_DEFAULT}

cd $KEEP_ECDSA_SOL_PATH
if [ "$NETWORK" != "alfajores" ]; then
  # Dafault app address.
  output=$(npx truffle exec scripts/get-default-application-account.js --network $NETWORK)
  CLIENT_APP_ADDRESS=$(echo "$output" | tail -1)
fi

if [ -z "$CLIENT_APP_ADDRESS" ]; then
  # Read user app address.
  read -p "Enter client application address: " client_app_address
  CLIENT_APP_ADDRESS=${client_app_address}
fi

if [ ! -z ${client_app_address+x} ]; then
  # Read user app when --application-address is set
  CLIENT_APP_ADDRESS=$client_app_address
fi

# Run script.
LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

printf "${LOG_START}Network:${LOG_END} $NETWORK"
printf "${LOG_START}Application address:${LOG_END} $CLIENT_APP_ADDRESS"

printf "${LOG_START}Starting initialization...${LOG_END}"

printf "${LOG_START}Configuring external client contract address...${LOG_END}"
CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    ./scripts/lcl-set-client-address.sh

printf "${LOG_START}Initializing contracts...${LOG_END}"
CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=$ACCOUNT_PRIVATE_KEY \
  npx truffle exec scripts/lcl-initialize.js --network $NETWORK

printf "${LOG_START}Updating keep-ecdsa config files...${LOG_END}"
for CONFIG_FILE in $KEEP_ECDSA_CONFIG_DIR_PATH/*.toml
do
  CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=$ACCOUNT_PRIVATE_KEY \
  KEEP_ECDSA_CONFIG_FILE_PATH=$CONFIG_FILE \
  CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    npx truffle exec scripts/lcl-client-config.js --network $NETWORK
done

printf "${DONE_START}Initialization completed!${DONE_END}"
