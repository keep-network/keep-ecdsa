#!/bin/bash

set -e

kubectl -n tbtc create configmap keep-ecdsa --from-file=files/
