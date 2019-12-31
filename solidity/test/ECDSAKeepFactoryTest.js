
import packTicket from './helpers/packTicket';
import generateTickets from './helpers/generateTickets';

const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBond = artifacts.require('KeepBond');

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory, tickets1, keepBond;
    const member = accounts[1]
    const operatorStakingWeight = 2000;
    const bondReference = 777;
    const bondAmount = 4242;

    describe("openKeep", async () => {
        beforeEach(async () => {
            keepBond = await KeepBond.new()
            keepFactory = await ECDSAKeepFactoryStub.new(keepBond.address)
        })

        //TODO: add snapshots

        it("emits ECDSAKeepCreated event upon keep creation", async () => {
            tickets1 = generateTickets(
                await keepFactory.getGroupSelectionRelayEntry(), 
                member, 
                operatorStakingWeight
            );

            let ticket = packTicket(tickets1[0].valueHex, 1, member);
            await keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: member});

            let blockNumber = await web3.eth.getBlockNumber()

            let keepAddress = await keepFactory.openKeep.call(
                10, // _groupSize
                5, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
            )

            await keepFactory.openKeep(
                10, // _groupSize
                5, // _honestThreshold
                "0xbc4862697a1099074168d54A555c4A60169c18BD" // _owner
            )

            let eventList = await keepFactory.getPastEvents('ECDSAKeepCreated', {
                fromBlock: blockNumber,
                toBlock: 'latest'
            })

            assert.isTrue(
                web3.utils.isAddress(keepAddress),
                `keep address ${keepAddress} is not a valid address`,
            );

            assert.equal(eventList.length, 1, "incorrect number of emitted events")

            assert.equal(
                eventList[0].returnValues.keepAddress,
                keepAddress,
                "incorrect keep address in emitted event",
            )

            assert.equal(
                eventList[0].returnValues.members,
                member,
                "incorrect keep member in emitted event",
            )
        })
    })
})
