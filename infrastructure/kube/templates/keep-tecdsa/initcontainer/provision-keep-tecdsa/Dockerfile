FROM node:11 AS runtime

WORKDIR /tmp

ARG CONTRACT_OWNER_ETH_ACCOUNT_ADDRESS
ENV CONTRACT_OWNER_ETH_ACCOUNT_ADDRESS=$CONTRACT_OWNER_ETH_ACCOUNT_ADDRESS

ARG ETH_HOSTNAME
ENV ETH_HOSTNAME=$ETH_HOSTNAME

ARG ETH_RPC_PORT
ENV ETH_RPC_PORT=$ETH_RPC_PORT

ARG ETH_WS_PORT
ENV ETH_WS_PORT=$ETH_WS_PORT

ARG ETH_NETWORK_ID
ENV ETH_NETWORK_ID=$ETH_NETWORK_ID

COPY ./BondedECDSAKeepFactory.json /tmp/BondedECDSAKeepFactory.json

COPY ./KeepBonding.json /tmp/KeepBonding.json

COPY ./TokenStaking.json /tmp/TokenStaking.json

COPY ./KeepToken.json /tmp/KeepToken.json

COPY ./keep-tecdsa-template.toml /tmp/keep-tecdsa-template.toml

COPY ./provision-keep-tecdsa.js /tmp/provision-keep-tecdsa.js

ENTRYPOINT ["node", "./provision-keep-tecdsa.js"]
