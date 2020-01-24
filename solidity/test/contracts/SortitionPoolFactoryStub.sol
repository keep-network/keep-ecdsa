pragma solidity ^0.5.10;

import "./SortitionPoolStub.sol";

contract SortitionPoolFactoryStub {

    function createSortitionPool() public payable returns(address) {
        return address(new SortitionPoolStub());
    }
}
