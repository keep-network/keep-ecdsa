#!/bin/bash
set -euo pipefail

LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)

# Defaults, can be overwritten by env variables/input parameters
NETWORK_DEFAULT="local"
CONFIG_DIR_PATH_DEFAULT=$(realpath -m $(dirname $0)/../configs)

help() {
  echo -e "\nUsage: ENV_VAR(S) $0" \
    "--config-dir <path>" \
    "--application-address <address>" \
    "--network <network>" \
    "--contracts-only"
  echo -e "\nEnvironment variables:\n"
  echo -e "\tCONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY: Contracts owner private key on Celo. Required for non-local network only"
  echo -e "\nCommand line arguments:\n"
  echo -e "\t--config-dir: Path to keep-ecdsa client configuration file(s)"
  echo -e "\t--application-address: Address of application approved by the operator"
  echo -e "\t--network: Celo network for keep-ecdsa client." \
    "Available networks and settings are specified in 'truffle.js'"
  echo -e "\t--contracts-only: Should execute contracts part only." \
    "Client installation will not be executed.\n"
  exit 1 # Exit script after printing help
}

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
  "--config-dir") set -- "$@" "-d" ;;
  "--application-address") set -- "$@" "-a" ;;
  "--network") set -- "$@" "-n" ;;
  "--contracts-only") set -- "$@" "-m" ;;
  "--help") set -- "$@" "-h" ;;
  *) set -- "$@" "$arg" ;;
  esac
done

# Parse short options
OPTIND=1
while getopts "d:a:n:mh" opt; do
  case "$opt" in
  d) config_dir_path="$OPTARG" ;;
  a) client_app_address="$OPTARG" ;;
  n) network="$OPTARG" ;;
  m) contracts_only=true ;;
  h) help ;;
  ?) help ;; # Print help in case parameter is non-existent
  esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

CONFIG_DIR_PATH=${config_dir_path:-$CONFIG_DIR_PATH_DEFAULT}
KEEP_ECDSA_CONFIG_DIR_PATH=$(realpath $CONFIG_DIR_PATH)
NETWORK=${network:-$NETWORK_DEFAULT}
CONTRACTS_ONLY=${contracts_only:-false}

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Linking dependencies...${LOG_END}"
npm link @keep-network/keep-core @keep-network/tbtc

# Default app address.
output=$(npx truffle exec scripts/get-default-application-account.js --network $NETWORK)
CLIENT_APP_ADDRESS=$(echo "$output" | tail -1)

if [ ! -z ${client_app_address+x} ]; then
  # Read user app when --application-address is set
  CLIENT_APP_ADDRESS=$client_app_address
fi

# Run script.
printf "${LOG_START}Network:${LOG_END} $NETWORK"
printf "${LOG_START}Application address:${LOG_END} $CLIENT_APP_ADDRESS"

printf "${LOG_START}Starting initialization...${LOG_END}"

if [ "$NETWORK" == "local" ]; then
  printf "${LOG_START}Initializing contracts...${LOG_END}"
  CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
    npx truffle exec scripts/lcl-initialize.js --network $NETWORK
fi

if [ "$CONTRACTS_ONLY" = false ]; then
  printf "${LOG_START}Updating keep-ecdsa config files...${LOG_END}"
  for CONFIG_FILE in $KEEP_ECDSA_CONFIG_DIR_PATH/*.toml; do
    KEEP_ECDSA_CONFIG_FILE_PATH=$CONFIG_FILE \
      CLIENT_APP_ADDRESS=$CLIENT_APP_ADDRESS \
      npx truffle exec scripts/lcl-client-config.js --network $NETWORK
  done
fi

printf "${DONE_START}Initialization completed!${DONE_END}"
