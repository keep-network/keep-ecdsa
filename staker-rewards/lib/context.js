import Web3 from "web3"
import ProviderEngine from "web3-provider-engine"
import WebsocketSubprovider from "web3-provider-engine/subproviders/websocket.js"
import Subproviders from "@0x/subproviders"

import Cache from "./cache.js"
import { Contract, getDeploymentBlockNumber } from "./contract-helper.js"
import Tenderly from "./tenderly.js"

import TokenStakingJson from "@keep-network/keep-core/artifacts/TokenStaking.json"
import BondedECDSAKeepFactoryJson from "@keep-network/keep-ecdsa/artifacts/BondedECDSAKeepFactory.json"
import BondedECDSAKeepJson from "@keep-network/keep-ecdsa/artifacts/BondedECDSAKeep.json"
import KeepBondingJson from "@keep-network/keep-ecdsa/artifacts/KeepBonding.json"
import BondedSortitionPoolJson from "@keep-network/keep-ecdsa/artifacts/BondedSortitionPool.json"

export const SANCTIONED_APPLICATION_ADDRESS =
  "0xe20A5C79b39bC8C363f0f49ADcFa82C2a01ab64a"

export default class Context {
  constructor(cache, web3, contracts, tenderly) {
    this.cache = cache
    this.web3 = web3
    this.contracts = contracts
    this.tenderly = tenderly
  }

  static async initialize(ethUrl, tenderlyProjectURL, tenderlyAccessToken) {
    const web3 = await initWeb3(ethUrl)

    const TokenStaking = new Contract(TokenStakingJson, web3)

    const BondedECDSAKeepFactory = new Contract(
      BondedECDSAKeepFactoryJson,
      web3
    )

    const BondedECDSAKeep = new Contract(BondedECDSAKeepJson, web3)

    const KeepBonding = new Contract(KeepBondingJson, web3)

    const BondedSortitionPool = new Contract(BondedSortitionPoolJson, web3)

    const factoryDeploymentBlock = await getDeploymentBlockNumber(
      BondedECDSAKeepFactoryJson,
      web3
    )

    const contracts = {
      TokenStaking: TokenStaking,
      BondedECDSAKeepFactory: BondedECDSAKeepFactory,
      BondedECDSAKeep: BondedECDSAKeep,
      KeepBonding: KeepBonding,
      BondedSortitionPool: BondedSortitionPool,
      sanctionedApplicationAddress: SANCTIONED_APPLICATION_ADDRESS,
      factoryDeploymentBlock: factoryDeploymentBlock,
    }

    const cache = new Cache(web3, contracts)
    await cache.initialize()

    let tenderly
    if (tenderlyAccessToken) {
      tenderly = Tenderly.initialize(
        web3,
        tenderlyProjectURL,
        tenderlyAccessToken
      )
    }

    return new Context(cache, web3, contracts, tenderly)
  }
}

async function initWeb3(url) {
  const engine = new ProviderEngine({ pollingInterval: 1000 })

  // MnemonicWalletSubprovider requires us to provide a valid mnemonic.
  // Since we are just reading, we provide a dummy mnemonic that has
  // no ether on it.
  const dummyMnemonic =
    "6892a90dab700bab8cee21cef939461f41f48b91c271120aa8b10cd3d9dd86dc"

  engine.addProvider(
    new Subproviders.MnemonicWalletSubprovider({ mnemonic: dummyMnemonic })
  )

  engine.addProvider(new WebsocketSubprovider({ rpcUrl: url }))

  const web3 = new Web3(engine)

  engine.start()

  // set the default account to the first account from the wallet
  web3.eth.defaultAccount = (await web3.eth.getAccounts())[0]

  return web3
}
