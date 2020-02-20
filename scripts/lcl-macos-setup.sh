#!/bin/bash
set -e

echo "Installing jq..."
brew install jq

echo "Installing truffle..."
npm install -g truffle
