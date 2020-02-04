pragma solidity ^0.5.10;

import "./BondedSortitionPoolStub.sol";

contract BondedSortitionPoolFactoryStub {
    function createSortitionPool() public payable returns (address) {
        return address(new BondedSortitionPoolStub());
    }
}
