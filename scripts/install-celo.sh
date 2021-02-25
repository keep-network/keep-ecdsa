#!/bin/bash
set -euo pipefail

LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

KEEP_CORE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../../keep-core)
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)

# Defaults, can be overwritten by env variables/input parameters
KEEP_CELO_PASSWORD=${KEEP_CELO_PASSWORD:-"password"}
NETWORK_DEFAULT="local"
CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY=${CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY:-""}

help()
{
   echo -e "\nUsage: ENV_VAR(S) $0"\
           "--keep-core-path <path>"\
           "--network <network>"
   echo -e "\nEnvironment variables:\n"
   echo -e "\tKEEP_CELO_PASSWORD: The password to unlock local Celo accounts to set up delegations."\
           "Required only for 'local' network. Default value is 'password'"
   echo -e "\tCONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY: Contracts owner private key on Celo. Required for non-local network only"
   echo -e "\nCommand line arguments:\n"
   echo -e "\t--keep-core-path: Path to the keep-core project"
   echo -e "\t--network: Celo network for keep-core client."\
           "Available networks and settings are specified in 'truffle.js'\n"
   exit 1 # Exit script after printing help
}

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--keep-core-path") set -- "$@" "-d" ;;
    "--network")        set -- "$@" "-n" ;;
    "--help")           set -- "$@" "-h" ;;
    *)                  set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "d:n:h" opt
do
   case "$opt" in
      d ) keep_core_path="$OPTARG" ;;
      n ) network="$OPTARG" ;;
      h ) help ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

KEEP_CORE_PATH=$(realpath ${keep_core_path:-$KEEP_CORE_PATH_DEFAULT})
NETWORK=${network:-$NETWORK_DEFAULT}

printf "${LOG_START}Path to the keep-core project: $KEEP_CORE_PATH ${LOG_END}"
printf "${LOG_START}Network: $NETWORK ${LOG_END}"

# Run script.
printf "${LOG_START}Starting installation...${LOG_END}"
KEEP_CORE_SOL_PATH=$(realpath $KEEP_CORE_PATH/solidity)
KEEP_CORE_SOL_ARTIFACTS_PATH=$(realpath $KEEP_CORE_SOL_PATH/build/contracts)

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install

if [ "$NETWORK" == "local" ]; then
    printf "${LOG_START}Unlocking celo accounts...${LOG_END}"
    KEEP_ETHEREUM_PASSWORD=$KEEP_CELO_PASSWORD \
        npx truffle exec scripts/unlock-eth-accounts.js --network $NETWORK
fi

printf "${LOG_START}Finding current celo network ID...${LOG_END}"
output=$(CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY=$CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY npx truffle exec ./scripts/get-network-id.js --network $NETWORK)
NETWORKID=$(echo "$output" | tail -1)
printf "Current network ID: ${NETWORKID}\n"

printf "${LOG_START}Fetching external contracts addresses...${LOG_END}"
KEEP_CORE_SOL_ARTIFACTS_PATH=$KEEP_CORE_SOL_ARTIFACTS_PATH \
NETWORKID=$NETWORKID \
    ./scripts/lcl-provision-external-contracts.sh

printf "${LOG_START}Migrating contracts...${LOG_END}"
npm run clean
CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY=$CONTRACT_OWNER_CELO_ACCOUNT_PRIVATE_KEY \
    npx truffle migrate --reset --network $NETWORK

printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
cd $KEEP_ECDSA_PATH
go generate ./...
go build -a -o keep-ecdsa .

printf "${DONE_START}Installation completed!${DONE_END}"
