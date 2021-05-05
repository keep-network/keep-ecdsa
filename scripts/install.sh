#!/bin/bash
set -euo pipefail

LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)

# Defaults, can be overwritten by env variables/input parameters
KEEP_ETHEREUM_PASSWORD=${KEEP_ETHEREUM_PASSWORD:-"password"}
NETWORK_DEFAULT="local"

help()
{
   echo -e "\nUsage: ENV_VAR(S) $0"\
           "--network <network>"
   echo -e "\nEnvironment variables:\n"
   echo -e "\tKEEP_ETHEREUM_PASSWORD: The password to unlock local Ethereum accounts to set up delegations."\
           "Required only for 'local' network. Default value is 'password'"
   echo -e "\nCommand line arguments:\n"
   echo -e "\t--network: Ethereum network for keep-ecdsa client."\
           "Available networks and settings are specified in 'truffle.js'\n"
   exit 1 # Exit script after printing help
}

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--network")        set -- "$@" "-n" ;;
    "--help")           set -- "$@" "-h" ;;
    *)                  set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "n:h" opt
do
   case "$opt" in
      n ) network="$OPTARG" ;;
      h ) help ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

# Overwrite default properties
NETWORK=${network:-$NETWORK_DEFAULT}

printf "${LOG_START}Network: $NETWORK ${LOG_END}"

# Run script.
printf "${LOG_START}Starting installation...${LOG_END}"

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install
npm link @keep-network/keep-core

if [ "$NETWORK" == "local" ]; then
  printf "${LOG_START}Unlocking ethereum accounts...${LOG_END}"
  KEEP_ETHEREUM_PASSWORD=$KEEP_ETHEREUM_PASSWORD \
      npx truffle exec scripts/unlock-eth-accounts.js --network $NETWORK
fi

printf "${LOG_START}Migrating contracts...${LOG_END}"
npm run clean
npx truffle migrate --reset --network $NETWORK

printf "${LOG_START}Creating links...${LOG_END}"
ln -sf build/contracts artifacts
npm link
printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
cd $KEEP_ECDSA_PATH
go generate ./...
go build -a -o keep-ecdsa .

printf "${DONE_START}Installation completed!${DONE_END}"
