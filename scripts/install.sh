#!/bin/bash
set -euo pipefail

GOPRIVATE="keep-network/keep-core"

LOG_START='\n\e[1;36m'  # new line + bold + cyan
LOG_END='\n\e[0m'       # new line + reset
DONE_START='\n\e[1;32m' # new line + bold + green
DONE_END='\n\n\e[0m'    # new line + reset

# Dafault inputs.
KEEP_ACCOUNT_PASSWORD_DEFAULT="password"
CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=""
NETWORK_DEFAULT="local"
KEEP_CORE_PATH_DEFAULT=$(realpath -m $(dirname $0)/../../keep-core)

help()
{
   echo ""
   echo "Usage: $0"\
        "--keep-core-path <path>"\
        "--account-password <password>"\
        "--private-key <private key>"\
        "--network <network>"
   echo -e "\t--keep-core-path: Path to the keep-core project"
   echo -e "\t--account-password: Account password"
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
    "--keep-core-path")    set -- "$@" "-d" ;;
    "--account-password")  set -- "$@" "-p" ;;
    "--private-key")       set -- "$@" "-k" ;;
    "--network")           set -- "$@" "-n" ;;
    *)                     set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "d:p:k:n:" opt
do
   case "$opt" in
      d ) keep_core_path="$OPTARG" ;;
      p ) account_password="$OPTARG" ;;
      k ) private_key="$OPTARG" ;;
      n ) network="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

KEEP_CORE_PATH=$(realpath ${keep_core_path:-$KEEP_CORE_PATH_DEFAULT})
KEEP_ACCOUNT_PASSWORD=${account_password:-$KEEP_ACCOUNT_PASSWORD_DEFAULT}
ACCOUNT_PRIVATE_KEY=${private_key:-$CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY}
NETWORK=${network:-$NETWORK_DEFAULT}

printf "${LOG_START}Path to the keep-core project: $KEEP_CORE_PATH ${LOG_END}"
printf "${LOG_START}Network: $NETWORK ${LOG_END}"

# Run script.
printf "${LOG_START}Starting installation...${LOG_END}"
KEEP_ECDSA_PATH=$(realpath $(dirname $0)/../)
KEEP_ECDSA_SOL_PATH=$(realpath $KEEP_ECDSA_PATH/solidity)
KEEP_CORE_SOL_PATH=$(realpath $KEEP_CORE_PATH/solidity)
KEEP_CORE_SOL_ARTIFACTS_PATH=$(realpath $KEEP_CORE_SOL_PATH/build/contracts)

cd $KEEP_ECDSA_SOL_PATH

printf "${LOG_START}Installing NPM dependencies...${LOG_END}"
npm install

if [ "$NETWORK" != "alfajores" ]; then
    printf "${LOG_START}Unlocking ethereum accounts...${LOG_END}"
    KEEP_ETHEREUM_PASSWORD=$KEEP_ACCOUNT_PASSWORD \
        npx truffle exec scripts/unlock-eth-accounts.js --network $NETWORK

fi

printf "${LOG_START}Finding current ethereum network ID...${LOG_END}"
output=$(CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=$ACCOUNT_PRIVATE_KEY npx truffle exec ./scripts/get-network-id.js --network $NETWORK)
NETWORKID=$(echo "$output" | tail -1)
printf "Current network ID: ${NETWORKID}\n"

printf "${LOG_START}Fetching external contracts addresses...${LOG_END}"
KEEP_CORE_SOL_ARTIFACTS_PATH=$KEEP_CORE_SOL_ARTIFACTS_PATH \
NETWORKID=$NETWORKID \
    ./scripts/lcl-provision-external-contracts.sh

printf "${LOG_START}Migrating contracts...${LOG_END}"
npm run clean
CONTRACT_OWNER_ACCOUNT_PRIVATE_KEY=$ACCOUNT_PRIVATE_KEY \
    npx truffle migrate --reset --network $NETWORK

printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
cd $KEEP_ECDSA_PATH
go generate ./...
go build -a -o keep-ecdsa .

printf "${DONE_START}Installation completed!${DONE_END}"
