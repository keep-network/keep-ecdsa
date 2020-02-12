#!/bin/bash

set -e

kubectl -n tbtc create configmap keep-tecdsa --from-file=files/
