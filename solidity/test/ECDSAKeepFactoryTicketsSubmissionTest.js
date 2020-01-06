import generateTickets from './helpers/generateTickets';
import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const { expectRevert } = require('openzeppelin-test-helpers');
const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBondingStub = artifacts.require('KeepBondingStub');
const BN = web3.utils.BN

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory, tickets1, keepBonding, 
    operator1 = accounts[2],
    operator2 = accounts[3];

    const operator1StakingWeight = 2000;
    const bondReference = 777;
    const bondAmount = new BN(50);
    const depositValue = new BN(100000)

    describe("ticket submission", async () => {
        before(async () => {
            keepBonding = await KeepBondingStub.new()
            keepFactory = await ECDSAKeepFactoryStub.new(keepBonding.address)

            await keepBonding.deposit(operator1, { value: depositValue })

            tickets1 = generateTickets(
                await keepFactory.getGroupSelectionRelayEntry(), 
                operator1, 
                operator1StakingWeight
            );
        })
        
        beforeEach(async () => {
            await createSnapshot()
        })

        afterEach(async () => {
            await restoreSnapshot()
        })

        it("should accept valid ticket with minimum virtual staker index", async () => {
            await keepFactory.submitTicket(tickets1[0], bondReference, bondAmount, {from: operator1});
        
            let submittedCount = await keepFactory.submittedTicketsCount();
            assert.equal(1, submittedCount, "Ticket should be accepted");
        });

        it("should accept valid ticket with maximum virtual staker index", async () => {
            await keepFactory.submitTicket(tickets1[tickets1.length - 1], bondReference, bondAmount, {from: operator1});
        
            let submittedCount = await keepFactory.submittedTicketsCount();
            assert.equal(1, submittedCount, "Ticket should be accepted");
        });

        it("should reject ticket with too high virtual staker index", async () => {
            let ticket = tickets1[tickets1.length - 1]
            ticket[31] = 209 // changing last byte for staker index, making it 2001

            await expectRevert(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Invalid ticket"
            );
        });

        it("should reject ticket with invalid value", async() => {
            let ticket = tickets1[0]
            ticket[0] = 19
            await expectRevert(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Invalid ticket"
            );
        });
        
        it("should reject ticket with not matching operator", async() => {
            await expectRevert(
                keepFactory.submitTicket(tickets1[0], bondReference, bondAmount, {from: operator2}),
                "Invalid ticket"
            )
        });
    
        it("should reject duplicate ticket", async () => {
            let ticket = tickets1[1];
            await keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1});
    
            await expectRevert(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Duplicate ticket"
            );
        });
    })

    // TODO: add more tests from keep-core when selectECDSAKeepMembers is implemented

})
