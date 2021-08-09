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

help() {
  echo -e "\nUsage: ENV_VAR(S) $0" \
    "--network <network>" \
    "--contracts-only"
  echo -e "\nEnvironment variables:\n"
  echo -e "\tKEEP_ETHEREUM_PASSWORD: The password to unlock local Ethereum accounts to set up delegations." \
    "Required only for 'local' network. Default value is 'password'"
  echo -e "\nCommand line arguments:\n"
  echo -e "\t--network: Ethereum network for keep-ecdsa client." \
    "Available networks and settings are specified in 'truffle.js'"
  echo -e "\t--contracts-only: Should execute contracts part only." \
    "Client installation will not be executed.\n"
  exit 1 # Exit script after printing help
}

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
  "--network") set -- "$@" "-n" ;;
  "--contracts-only") set -- "$@" "-m" ;;
  "--help") set -- "$@" "-h" ;;
  *) set -- "$@" "$arg" ;;
  esac
done

# Parse short options
OPTIND=1
while getopts "n:mh" opt; do
  case "$opt" in
  n) network="$OPTARG" ;;
  m) contracts_only=true ;;
  h) help ;;
  ?) help ;; # Print help in case parameter is non-existent
  esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

# Overwrite default properties
NETWORK=${network:-$NETWORK_DEFAULT}
CONTRACTS_ONLY=${contracts_only:-false}

# Run script.
printf "${LOG_START}Network: $NETWORK ${LOG_END}"
printf "${LOG_START}Starting installation...${LOG_END}"
printf "${LOG_START}Installing root NPM dependencies...${LOG_END}"

cd $KEEP_ECDSA_PATH
npm install

printf "${LOG_START}Installing Solidity NPM dependencies...${LOG_END}"

cd $KEEP_ECDSA_SOL_PATH
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

printf "${LOG_START}Copying contract artifacts...${LOG_END}"
rm -rf artifacts
cp -r build/contracts artifacts
npm link
npm link @keep-network/keep-core

if [ "$CONTRACTS_ONLY" = false ]; then
  printf "${LOG_START}Building keep-ecdsa client...${LOG_END}"
  cd $KEEP_ECDSA_PATH

  # solc doesn't support symbolic links that are made in `node_modules` by `npm link`
  # command. We need to update the `--allow-paths` value to be the parent directory
  # that is assumed to contain both current project and dependent project.
  # Ref: https://github.com/ethereum/solidity/issues/4623
  TMP_FILE=$(mktemp /tmp/Makefile-ethereum.XXXXXXXXXX)
  sed 's/--allow-paths ${solidity_dir}/--allow-paths $(realpath ${SOLIDITY_DIR}\/..\/..\/)/g' pkg/chain/gen/ethereum/Makefile >$TMP_FILE
  mv $TMP_FILE pkg/chain/gen/ethereum/Makefile
  TMP_FILE=$(mktemp /tmp/Makefile-celo.XXXXXXXXXX)
  sed 's/--allow-paths ${solidity_dir}/--allow-paths $(realpath ${SOLIDITY_DIR}\/..\/..\/)/g' pkg/chain/gen/celo/Makefile >$TMP_FILE
  mv $TMP_FILE pkg/chain/gen/celo/Makefile

  go generate ./...
  go build -a -o keep-ecdsa .
fi

printf "${DONE_START}Installation completed!${DONE_END}"
