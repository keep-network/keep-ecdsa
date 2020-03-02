import { increaseTime } from './helpers/increaseTime';
const { expectRevert } = require('openzeppelin-test-helpers');

const Registry = artifacts.require('Registry');
const BondedECDSAKeepFactoryStub = artifacts.require('BondedECDSAKeepFactoryStub');
const KeepBonding = artifacts.require('KeepBonding');
const TokenStakingStub = artifacts.require("TokenStakingStub")
const BondedSortitionPoolFactory = artifacts.require('BondedSortitionPoolFactory');
const RandomBeaconStub = artifacts.require('RandomBeaconStub')


const BondedECDSAKeep = artifacts.require('BondedECDSAKeep');
const ECDSAKeepRewards = artifacts.require('ECDSAKeepRewards');

contract('ECDSAKeepRewards', (accounts) => {
    let masterKeep
    let factory 
    let registry
    let rewards

    let tokenStaking
    let keepFactory
    let bondedSortitionPoolFactory
    let keepBonding
    let randomBeacon
    let signerPool

    async function initializeNewFactory() {
        registry = await Registry.new()
        bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
        tokenStaking = await TokenStakingStub.new()
        keepBonding = await KeepBonding.new(registry.address, tokenStaking.address)
        randomBeacon = await RandomBeaconStub.new()
        const bondedECDSAKeepMasterContract = await BondedECDSAKeep.new()
        keepFactory = await BondedECDSAKeepFactoryStub.new(
            bondedECDSAKeepMasterContract.address,
            bondedSortitionPoolFactory.address,
            tokenStaking.address,
            keepBonding.address,
            randomBeacon.address
        )
    }
  before(async () => {
    await initializeNewFactory()

    rewards = await ECDSAKeepRewards.new(0, 0, accounts[0], 0, keepFactory.address)
    })

  describe("find", async () => {
    let block
    let increment = 1000
    let count = 15
    let timestamps = []
    before(async () => {
    // creates 100 keeps
    block = await web3.eth.getBlock("latest")

     for (let i=0;i<count;i++){
        await keepFactory.stubOpenKeep(accounts[0])
        const keepAddress = await keepFactory.getKeepAtIndex(i)
        const keepContract = await BondedECDSAKeep.at(keepAddress)
        const timestamp = await keepContract.getTimestamp.call()
        timestamps.push(timestamp.toString())
        increaseTime(increment)
    }
    })

    it("finds correct values - unbounded", async () => {
      for(let i = 0; i < count - 1; i++){

        let expectedIndex = i
        let targetTimestamp = Number(timestamps[expectedIndex]) + 1
        let index = await rewards.find(0, count, targetTimestamp)

        expect(index.toNumber()).to.equal(expectedIndex)
      } 
    })

    it("finds correct values - bounded", async () => {
      const upperBound = 14
      const lowerBound = 4
      for(let i = lowerBound; i < upperBound - 1 ; i++){
        let expectedIndex = i
        let targetTimestamp = Number(timestamps[expectedIndex]) + 1
        let index = await rewards.find(4, 14, targetTimestamp)
        expect(index.toNumber()).to.equal(expectedIndex)
      } 
    })

    it("reverts if out of bounds", async () => {
      let badIndex = 9
      let targetTimestamp = Number(timestamps[badIndex]) + 1
      await expectRevert(
        rewards.find(10, 14, targetTimestamp),
        "could not find target"
      )
    })
  })
})
