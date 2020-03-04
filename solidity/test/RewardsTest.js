import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";
import { increaseTime } from './helpers/increaseTime';
const { expectRevert } = require('openzeppelin-test-helpers');

const Registry = artifacts.require('Registry');
const RewardsFactoryStub = artifacts.require('RewardsFactoryStub');
const KeepBonding = artifacts.require('KeepBonding');
const TokenStakingStub = artifacts.require("TokenStakingStub")
const BondedSortitionPoolFactory = artifacts.require('BondedSortitionPoolFactory');
const RandomBeaconStub = artifacts.require('RandomBeaconStub')


const RewardsKeepStub = artifacts.require('RewardsKeepStub');
const ECDSAKeepRewards = artifacts.require('ECDSAKeepRewards');

contract.only('ECDSAKeepRewards', (accounts) => {
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

    // defaultTimestamps[i] == 1000 + i
    const defaultTimestamps = [
        1000,
        1001,
        1002,
        1003,
        1004,
        1005,
        1006,
        1007,
        1008,
        1009,
        1010,
        1011,
        1012,
        1013,
        1014,
        1015,
    ]

    const initiationTime = 1000
    const termLength = 100

    async function initializeNewFactory() {
        registry = await Registry.new()
        bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
        tokenStaking = await TokenStakingStub.new()
        keepBonding = await KeepBonding.new(registry.address, tokenStaking.address)
        randomBeacon = await RandomBeaconStub.new()
        const bondedECDSAKeepMasterContract = await RewardsKeepStub.new()
        keepFactory = await RewardsFactoryStub.new(
            bondedECDSAKeepMasterContract.address,
            bondedSortitionPoolFactory.address,
            tokenStaking.address,
            keepBonding.address,
            randomBeacon.address
        )
    }

    async function createKeeps(timestamps) {
        await keepFactory.openSyntheticKeeps(timestamps)
        for (let i = 0; i < timestamps.length; i++) {
            let keepAddress = await keepFactory.getKeepAtIndex(i)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.setTimestamp(timestamps[i])
        }
    }

    before(async () => {
        await initializeNewFactory()

        rewards = await ECDSAKeepRewards.new(
            termLength,
            0,
            accounts[0],
            0,
            keepFactory.address,
            initiationTime
        )
    })

    beforeEach(async () => {
        await createSnapshot()
    })

    afterEach(async () => {
        await restoreSnapshot()
    })

    describe("intervalOf", async () => {
        it("returns the correct interval", async () => {
            let interval1000 = await rewards.intervalOf(1000)
            expect(interval1000.toNumber()).to.equal(0)
            let interval1001 = await rewards.intervalOf(1001)
            expect(interval1001.toNumber()).to.equal(0)
            let interval1099 = await rewards.intervalOf(1099)
            expect(interval1099.toNumber()).to.equal(0)
            let interval1100 = await rewards.intervalOf(1100)
            expect(interval1100.toNumber()).to.equal(1)
            let interval1101 = await rewards.intervalOf(1101)
            expect(interval1101.toNumber()).to.equal(1)
            let interval1000000 = await rewards.intervalOf(1000000)
            expect(interval1000000.toNumber()).to.equal(9990)
        })

        it("reverts on timestamps before initiation", async () => {
            await expectRevert(
                rewards.intervalOf(999),
                "Timestamp is before the first interval"
            )
        })
    })

    describe("startOf", async () => {
        it("returns the start of the interval", async () => {
            let end0 = await rewards.startOf(0)
            expect(end0.toNumber()).to.equal(1000)
            let end1 = await rewards.startOf(1)
            expect(end1.toNumber()).to.equal(1100)
            let end9990 = await rewards.startOf(9990)
            expect(end9990.toNumber()).to.equal(1000000)
        })
    })

    describe("endOf", async () => {
        it("returns the end of the interval", async () => {
            let end0 = await rewards.endOf(0)
            expect(end0.toNumber()).to.equal(1100)
            let end1 = await rewards.endOf(1)
            expect(end1.toNumber()).to.equal(1200)
            let end9990 = await rewards.endOf(9990)
            expect(end9990.toNumber()).to.equal(1000100)
        })
    })

    describe("findEndpoint", async () => {
        let increment = 1000

        it("returns 0 when no keeps have been created", async () => {
            let targetTimestamp = await rewards.currentTime()
            increaseTime(increment)

            let index = await rewards.findEndpoint(targetTimestamp)
            expect(index.toNumber()).to.equal(0)
        })

        it("reverts if the endpoint is in the future", async () => {
            let recentTimestamp = await rewards.currentTime()
            let targetTimestamp = recentTimestamp + increment
            await expectRevert(
                rewards.findEndpoint(targetTimestamp),
                "interval hasn't ended yet"
            )
        })

        it("returns the first index outside the interval", async () => {
            let timestamps = defaultTimestamps
            await createKeeps(timestamps)
            for (let i = 0; i < 16; i++) {
                let expectedIndex = i
                let targetTimestamp = timestamps[i]
                let index = await rewards.findEndpoint(targetTimestamp)

                expect(index.toNumber()).to.equal(expectedIndex)
            }
        })

        it("returns the next keep's index when all current keeps were created in the interval", async () => {
            let timestamps = defaultTimestamps
            await createKeeps(timestamps)
            let targetTimestamp = 2000
            let expectedIndex = 16
            let index = await rewards.findEndpoint(targetTimestamp)

            expect(index.toNumber()).to.equal(expectedIndex)
        })

        it("returns 0 when all current keeps were created after the interval", async () => {
            let timestamps = defaultTimestamps
            await createKeeps(timestamps)
            let targetTimestamp = 500
            let expectedIndex = 0
            let index = await rewards.findEndpoint(targetTimestamp)

            expect(index.toNumber()).to.equal(expectedIndex)
        })

        it("returns the correct index when duplicates are present", async () => {
            let timestamps = [1001, 1001, 1002, 1002]
            await createKeeps(timestamps)
            let targetTimestamp = 1002
            let expectedIndex = 2
            let index = await rewards.findEndpoint(targetTimestamp)

            expect(index.toNumber()).to.equal(expectedIndex)
        })
    })
})
