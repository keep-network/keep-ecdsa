pragma solidity 0.5.17;

contract ManagedGrantStub {
    address public grantee;

    constructor(address _grantee) public {
        grantee = _grantee;
    }
}