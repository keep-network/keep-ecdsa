const {accounts, contract, web3} = require("@openzeppelin/test-environment")
const { createSnapshot, restoreSnapshot } = require("./helpers/snapshot");
const { increaseTime } = require('./helpers/increaseTime');
const { expectRevert } = require('@openzeppelin/test-helpers');

const StackLib = contract.fromArtifact("StackLib")
const Registry = contract.fromArtifact('KeepRegistry');
const RewardsFactoryStub = contract.fromArtifact('RewardsFactoryStub');
const KeepBonding = contract.fromArtifact('KeepBonding');
const TokenStakingStub = contract.fromArtifact("TokenStakingStub")
const TokenGrantStub = contract.fromArtifact("TokenGrantStub")
const BondedSortitionPoolFactory = contract.fromArtifact('BondedSortitionPoolFactory');
const RandomBeaconStub = contract.fromArtifact('RandomBeaconStub')
const TestToken = contract.fromArtifact('TestToken')

const RewardsKeepStub = contract.fromArtifact('RewardsKeepStub');
const ECDSAKeepRewards = contract.fromArtifact('ECDSAKeepRewardsStub');

const chai = require("chai")
const expect = chai.expect
const assert = chai.assert

describe.only('ECDSAKeepRewards', () => {
    const alice = accounts[0]
    const bob = accounts[1]
    const aliceBeneficiary = accounts[2]
    const bobBeneficiary = accounts[3]
    const funder = accounts[9]

    let registry
    let rewards
    let token

    let tokenStaking
    let tokenGrant
    let keepFactory
    let bondedSortitionPoolFactory
    let keepBonding
    let randomBeacon

    const rewardTimestamps = [
        1000, 1001, 1099, // interval 0; 0..2
        1100, 1101, 1102, 1103, // interval 1; 3..6
        1234, // interval 2; 7
        1300, 1301, // interval 3; 8..9
        1500, // interval 5; 10
        1600, 1601, // interval 6; 11..12
    ]

    const intervalWeights = [
        // percentage of unallocated rewards, allocated : remaining
        20, // 20:80
        50, // 40:40
        25, // 10:30
        50, // 15:15
    ]

    const initiationTime = 1000
    const termLength = 100
    const totalRewards = 1000000
    const minimumIntervalKeeps = 2


    async function initializeNewFactory() {
        registry = await Registry.new()
        bondedSortitionPoolFactory = await BondedSortitionPoolFactory.new()
        tokenStaking = await TokenStakingStub.new()
        tokenGrant = await TokenGrantStub.new()
        keepBonding = await KeepBonding.new(registry.address, tokenStaking.address, tokenGrant.address)
        randomBeacon = await RandomBeaconStub.new()
        const bondedECDSAKeepMasterContract = await RewardsKeepStub.new()
        keepFactory = await RewardsFactoryStub.new(
            bondedECDSAKeepMasterContract.address,
            bondedSortitionPoolFactory.address,
            tokenStaking.address,
            keepBonding.address,
            randomBeacon.address
        )

        await tokenStaking.setBeneficiary(alice, aliceBeneficiary)
        await tokenStaking.setBeneficiary(bob, bobBeneficiary)
    }

    async function createKeeps(timestamps) {
        await keepFactory.openSyntheticKeeps([alice, bob], timestamps)
        for (let i = 0; i < timestamps.length; i++) {
            let keepAddress = await keepFactory.getKeepAtIndex(i)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.setTimestamp(timestamps[i])
        }
    }

    async function fund(amount) {
        await token.mint(funder, amount)
        await token.approveAndCall(
            rewards.address,
            amount,
            "0x0",
            { from: funder }
        )
    }

    before(async () => {
        await BondedSortitionPoolFactory.detectNetwork()
        await BondedSortitionPoolFactory.link(
            "StackLib",
            (await StackLib.new()).address
        )
        await initializeNewFactory()

        token = await TestToken.new()

        rewards = await ECDSAKeepRewards.new(
            termLength,
            token.address,
            minimumIntervalKeeps,
            keepFactory.address,
            initiationTime,
            intervalWeights
        )

        await fund(totalRewards)
    })

    beforeEach(async () => {
        await createSnapshot()
    })

    afterEach(async () => {
        await restoreSnapshot()
    })

    describe("bytes32/address conversions", async () => {
        it("converts to address", async () => {
            let inputA = "0x111122223333444455556666777788889999AAAABBBBCCCCDDDDEEEEFFFFCCCC"
            let inputB = "0x111122223333444455556666777788889999aaaa000000000000000000000000"
            let outputA = await rewards._toAddress(inputA);
            let outputB = await rewards._toAddress(inputB);
            expect(outputA).to.equal("0x111122223333444455556666777788889999aAaa")
            expect(outputB).to.equal("0x111122223333444455556666777788889999aAaa")
        })

        it("converts to bytes32", async () => {
            let input = "0x111122223333444455556666777788889999aAaa"
            let output = "0x111122223333444455556666777788889999aaaa000000000000000000000000"
            expect(await rewards._fromAddress(input)).to.equal(output)
        })

        it("checks validity", async () => {
            let inputA = "0x111122223333444455556666777788889999AAAABBBBCCCCDDDDEEEEFFFFCCCC"
            let inputB = "0x111122223333444455556666777788889999aaaa000000000000000000000000"
            let inputC = "0x111122223333444455556666777788889999aaaa000000000000000000000001"
            expect(await rewards.testRoundtrip(inputA)).to.equal(false)
            expect(await rewards.testRoundtrip(inputB)).to.equal(true)
            expect(await rewards.testRoundtrip(inputC)).to.equal(false)
        })
    })

    describe("eligibleForReward", async () => {
        it("returns true for happily closed keeps", async () => {
            await createKeeps([1000])
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.close()
            console.log(keepAddress)
            console.log(await rewards._toAddress(keepAddress));
            let addressBytes = await rewards._fromAddress(keepAddress)
            console.log(addressBytes)
            console.log(await rewards._toAddress(addressBytes));
            await rewards.testRoundtrip(keepAddress)
            let eligible = await rewards.eligibleForReward(keepAddress)
            expect(eligible).to.equal(true)
        })

        it("returns false for terminated keeps", async () => {
            await createKeeps([1000])
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.terminate()
            let eligible = await rewards.eligibleForReward(keepAddress)
            expect(eligible).to.equal(false)
        })

        it("returns false for active keeps", async () => {
            await createKeeps([1000])
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let eligible = await rewards.eligibleForReward(keepAddress)
            expect(eligible).to.equal(false)
        })
    })


    describe("receiveReward", async () => {
        it("lets closed keeps claim the reward correctly", async () => {
            let timestamps = rewardTimestamps
            await createKeeps(timestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.close()
            await rewards.receiveReward(keepAddress)
            let aliceBalance = await token.balanceOf(aliceBeneficiary)
            expect(aliceBalance.toNumber()).to.equal(33333)
        })

        it("doesn't let keeps claim rewards again", async () => {
            let timestamps = rewardTimestamps
            await createKeeps(timestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.close()
            await rewards.receiveReward(keepAddress)
            await expectRevert(
                rewards.receiveReward(keepAddress),
                "Rewards already claimed"
            )
        })

        it("doesn't let active keeps claim the reward", async () => {
            await createKeeps(rewardTimestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            await expectRevert(
                rewards.receiveReward(keepAddress),
                "Keep is not closed"
            )
        })

        it("doesn't let terminated keeps claim the reward", async () => {
            await createKeeps(rewardTimestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.terminate()
            await expectRevert(
                rewards.receiveReward(keepAddress),
                "Keep is not closed"
            )
        })

        it("doesn't let unrecognized keeps claim the reward", async () => {
            await createKeeps(rewardTimestamps)
            let fakeKeepAddress = accounts[8]
            await expectRevert(
                rewards.receiveReward(fakeKeepAddress),
                "Keep address not recognized by factory"
            )
        })

        it("requires that the interval is over", async () => {
            let recentTimestamp = await rewards.currentTime()
            let targetTimestamp = recentTimestamp + 1000
            await createKeeps([targetTimestamp])
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.close()
            await expectRevert(
                rewards.receiveReward(keepAddress),
                "Interval hasn't ended yet"
            )
        })
    })

    describe("reportTermination", async () => {
        it("unallocates rewards allocated to terminated keeps", async () => {
            let timestamps = rewardTimestamps
            await createKeeps(timestamps)

            let closedKeepAddress = await keepFactory.getKeepAtIndex(1)
            let closedKeep = await RewardsKeepStub.at(closedKeepAddress)
            await closedKeep.close()
            await rewards.receiveReward(closedKeepAddress) // allocate rewards

            let terminatedKeepAddress = await keepFactory.getKeepAtIndex(0)
            let terminatedKeep = await RewardsKeepStub.at(terminatedKeepAddress)
            await terminatedKeep.terminate()
            let preUnallocated = await rewards.getUnallocatedRewards()
            await rewards.reportTermination(terminatedKeepAddress)
            let postUnallocated = await rewards.getUnallocatedRewards()
            expect(postUnallocated.toNumber()).to.equal(
                preUnallocated.toNumber() + 66666
            )
        })

        it("doesn't unallocate rewards twice for the same keep", async () => {
            let timestamps = rewardTimestamps
            await createKeeps(timestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.terminate()
            await rewards.reportTermination(keepAddress)
            await expectRevert(
                rewards.reportTermination(keepAddress),
                "Rewards already claimed"
            )
        })

        it("doesn't unallocate active keeps' rewards", async () => {
            await createKeeps(rewardTimestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            await expectRevert(
                rewards.reportTermination(keepAddress),
                "Keep is not terminated"
            )
        })

        it("doesn't unallocate closed keeps' rewards", async () => {
            await createKeeps(rewardTimestamps)
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.close()
            await expectRevert(
                rewards.reportTermination(keepAddress),
                "Keep is not terminated"
            )
        })

        it("doesn't unallocate unrecognized keeps' rewards", async () => {
            await createKeeps(rewardTimestamps)
            let fakeKeepAddress = accounts[8]
            await expectRevert(
                rewards.reportTermination(fakeKeepAddress),
                "Keep address not recognized by factory"
            )
        })

        it("requires that the interval is over", async () => {
            let recentTimestamp = await rewards.currentTime()
            let targetTimestamp = recentTimestamp + 1000
            await createKeeps([targetTimestamp])
            let keepAddress = await keepFactory.getKeepAtIndex(0)
            let keep = await RewardsKeepStub.at(keepAddress)
            await keep.terminate()
            await expectRevert(
                rewards.reportTermination(keepAddress),
                "Interval hasn't ended yet"
            )
        })
    })
})
