pragma solidity ^0.5.4;

import "../../contracts/ECDSAKeepFactory.sol";

/// @title ECDSA Keep Factory Tickets Ordering Stub
/// @dev This contract is for testing purposes only.
contract ECDSAKeepFactoryTicketsOrderingStub is ECDSAKeepFactory {

    constructor(address _keepBondContract)
        ECDSAKeepFactory(_keepBondContract) public {}

    /**
    * @dev Gets an index of a ticket that a higherTicketValueIndex points to.
    * Ex. tickets[23, 5, 65]
    * getPreviousTicketIndex(2) = 0
    */
    function getPreviousTicketIndex(uint256 higherTicketValueIndex) public view returns (uint256) {
        return previousTicketIndex[higherTicketValueIndex];
    }

    function setGroupSize(uint256 size) public {
        groupSize = size;
    }

    function addNewTicket(uint64 newTicketValue) public {
        addTicket(newTicketValue);
    }

     /**
    * @dev Gets submitted group candidate tickets so far.
    */
    function getAllTickets() public view returns (uint64[] memory) {
        return tickets;
    }

    /**
    * @dev Gets an index of the highest ticket value (tail).
    */
    function getTail() public view returns (uint256) {
        return tail;
    }
}
