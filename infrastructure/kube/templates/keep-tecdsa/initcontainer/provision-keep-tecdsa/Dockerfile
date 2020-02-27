FROM node:11 AS runtime

WORKDIR /tmp

COPY ./package.json /tmp/package.json
COPY ./package-lock.json /tmp/package-lock.json

RUN npm install

COPY ./BondedECDSAKeepFactory.json /tmp/BondedECDSAKeepFactory.json

COPY ./KeepBonding.json /tmp/KeepBonding.json

COPY ./TokenStaking.json /tmp/TokenStaking.json

COPY ./KeepToken.json /tmp/KeepToken.json

COPY ./keep-tecdsa-config-template.toml /tmp/keep-tecdsa-config-template.toml

COPY ./provision-keep-tecdsa.js /tmp/provision-keep-tecdsa.js

ENTRYPOINT ["node", "./provision-keep-tecdsa.js"]
