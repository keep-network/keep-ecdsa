FROM node:11 AS runtime

WORKDIR /tmp

COPY ./package.json /tmp/package.json
COPY ./package-lock.json /tmp/package-lock.json

RUN npm install

COPY ./BondedECDSAKeepFactory.json /tmp/BondedECDSAKeepFactory.json

COPY ./KeepBonding.json /tmp/KeepBonding.json

COPY ./TokenStaking.json /tmp/TokenStaking.json

COPY ./KeepToken.json /tmp/KeepToken.json

COPY ./TBTCSystem.json /tmp/TBTCSystem.json

COPY ./keep-ecdsa-config-template.toml /tmp/keep-ecdsa-config-template.toml

COPY ./provision-keep-ecdsa.js /tmp/provision-keep-ecdsa.js

ENTRYPOINT ["node", "./provision-keep-ecdsa.js"]
