import packTicket from './helpers/packTicket';
import generateTickets from './helpers/generateTickets';
import expectThrowWithMessage from './helpers/expectThrowWithMessage';
import { createSnapshot, restoreSnapshot } from "./helpers/snapshot";

const ECDSAKeepFactoryStub = artifacts.require('ECDSAKeepFactoryStub');
const KeepBonding = artifacts.require('KeepBonding');
const BN = web3.utils.BN

const chai = require('chai')
chai.use(require('bn-chai')(BN))

contract("ECDSAKeepFactory", async accounts => {
    let keepFactory, tickets1, keepBonding,
    operator1 = accounts[1],
    operator2 = accounts[2];

    const operator1StakingWeight = 2000;
    const bondReference = 777;
    const bondAmount = new BN(50);
    const depositValue = new BN(100000)

    describe("ticket submission", async () => {
        before(async () => {
            keepBonding = await KeepBonding.new()
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
            let ticket = packTicket(tickets1[0].valueHex, 1, operator1);
            await keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1});
        
            let submittedCount = await keepFactory.submittedTicketsCount();
            assert.equal(1, submittedCount, "Ticket should be accepted");
        });

        it("should accept valid ticket with maximum virtual staker index", async () => {
            let ticket = packTicket(tickets1[tickets1.length - 1].valueHex, tickets1.length, operator1);
            await keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1});
        
            let submittedCount = await keepFactory.submittedTicketsCount();
            assert.equal(1, submittedCount, "Ticket should be accepted");
        });

        it("should reject ticket with too high virtual staker index", async () => {
            let ticket = packTicket(tickets1[tickets1.length - 1].valueHex, tickets1.length + 1, operator1);
            await expectThrowWithMessage(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Invalid ticket"
            );
        });

        it("should reject ticket with invalid value", async() => {
            let ticket = packTicket('0x1337', 1, operator1);
            await expectThrowWithMessage(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Invalid ticket"
            );
        });
        
        it("should reject ticket with not matching operator", async() => {
            let ticket = packTicket(tickets1[0].valueHex, 1, operator1);
            await expectThrowWithMessage(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator2}),
                "Invalid ticket"
            )
        });
    
        it("should reject ticket with not matching virtual staker index", async() => {
            let ticket = packTicket(tickets1[0].valueHex, 2, operator1);
            await expectThrowWithMessage(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Invalid ticket"
            )
        });
    
        it("should reject duplicate ticket", async () => {
            let ticket = packTicket(tickets1[0].valueHex, 1, operator1);
            await keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1});
    
            await expectThrowWithMessage(
                keepFactory.submitTicket(ticket, bondReference, bondAmount, {from: operator1}),
                "Duplicate ticket"
            );
        });
    })

    // TODO: add more tests from keep-core when selectECDSAKeepMembers is implemented

})
