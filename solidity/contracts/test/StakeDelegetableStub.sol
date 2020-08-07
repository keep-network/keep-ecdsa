pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/StakeDelegatable.sol";

/// @title Stake Delegatable Stub
/// @dev This contract is for testing purposes only.
contract StakeDelegatableStub is StakeDelegatable {
    mapping(address => uint256) stakes;

    mapping(address => address) operatorToOwner;
    mapping(address => address payable) operatorToBeneficiary;
    mapping(address => address) operatorToAuthorizer;

    function setBalance(address _operator, uint256 _balance) public {
        stakes[_operator] = _balance;
    }

    function balanceOf(address _address) public view returns (uint256 balance) {
        return stakes[_address];
    }

    function setOwner(address _operator, address _owner) public {
        operatorToOwner[_operator] = _owner;
    }

    function ownerOf(address _operator) public view returns (address) {
        return operatorToOwner[_operator];
    }

    function setBeneficiary(address _operator, address payable _beneficiary)
        public
    {
        operatorToBeneficiary[_operator] = _beneficiary;
    }

    function beneficiaryOf(address _operator)
        public
        view
        returns (address payable)
    {
        return operatorToBeneficiary[_operator];
    }

    function setAuthorizer(address _operator, address _authorizer) public {
        operatorToAuthorizer[_operator] = _authorizer;
    }

    function authorizerOf(address _operator) public view returns (address) {
        return operatorToAuthorizer[_operator];
    }
}
