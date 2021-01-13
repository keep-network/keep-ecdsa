#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

help()
{
   echo ""
   echo "Usage: $0 --etherscan-token <etherscan_token> --timestamp <timestamp>"
   echo -e "\t--etherscan-token Etherscan API key"
   echo -e "\t--timestamp Timestamp of the searched block"
   exit 1 # Exit script after printing help
}

if [ "$1" == "-help" ]; then
  help
fi

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--etherscan-token") set -- "$@" "-e" ;;
    "--timestamp")       set -- "$@" "-t" ;;
    *)                   set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "e:t:" opt
do
   case "$opt" in
      e ) etherscan_token="$OPTARG" ;;
      t ) timestamp="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

#Print help in case required parameters are empty
if [ -z "$etherscan_token" ] || [ -z "$timestamp" ]
then
   echo "Some or all of the required parameters are empty";
   help
fi

url="https://api.etherscan.io/api?\
module=block&\
action=getblocknobytime&\
timestamp=$timestamp&\
closest=after&\
apikey=$etherscan_token"

block=$(curl -s $url | jq '.result|tonumber')

printf "${LOG_START}Block $block${LOG_END}"