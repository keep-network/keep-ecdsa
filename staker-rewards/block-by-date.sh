#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

WORKDIR=$PWD

NODE_CURRENT_VER="$(node --version)"
NODE_REQUIRED_VER="v14.3.0"

if [ "$(printf '%s\n' "$NODE_REQUIRED_VER" "$NODE_CURRENT_VER" | sort -V | head -n1)" != "$NODE_REQUIRED_VER" ]; 
then
      echo "Required node version must be at least ${NODE_REQUIRED_VER}" 
      exit 1
fi

help()
{
   echo ""
   echo "Usage: $0 --eth-host <eth_host> --timestamp <timestamp>"
   echo -e "\t--eth-host Websocket endpoint of the Ethereum node"
   echo -e "\t--timestamp Timestamp of the searched block"
   exit 1 # Exit script after printing help
}

if [ "$1" == "-help" ]; then
  help
fi

printf "${LOG_START}Processing input parameters...${LOG_END}"

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--eth-host")  set -- "$@" "-h" ;;
    "--timestamp") set -- "$@" "-t" ;;
    *)             set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "h:t:" opt
do
   case "$opt" in
      h ) eth_host="$OPTARG" ;;
      t ) timestamp="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

#Print help in case required parameters are empty
if [ -z "$eth_host" ] || [ -z "$timestamp" ]
then
   echo "Some or all of the required parameters are empty";
   help
fi

printf "${LOG_START}Installing dependencies for staker-rewards...${LOG_END}"

cd "$WORKDIR"
npm i

printf "${LOG_START}Looking for block...${LOG_END}"

ETH_HOSTNAME="$eth_host" node --experimental-json-modules block-by-date.js "$timestamp"

printf "${LOG_START}Script finished successfully${LOG_END}"