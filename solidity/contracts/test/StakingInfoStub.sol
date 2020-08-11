pragma solidity 0.5.17;

/// @title Staking Stub
/// @dev This contract is for testing purposes only.
contract StakingInfoStub {
    // Authorized operator contracts.
    mapping(address => mapping(address => bool)) internal authorizations;

    mapping(address => address) operatorToOwner;
    mapping(address => address payable) operatorToBeneficiary;
    mapping(address => address) operatorToAuthorizer;

    function authorizeOperatorContract(
        address _operator,
        address _operatorContract
    ) public {
        authorizations[_operatorContract][_operator] = true;
    }

    function isAuthorizedForOperator(
        address _operator,
        address _operatorContract
    ) public view returns (bool) {
        return authorizations[_operatorContract][_operator];
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

    function setOwner(address _operator, address _owner) public {
        operatorToOwner[_operator] = _owner;
    }

    function ownerOf(address _operator) public view returns (address) {
        return operatorToOwner[_operator];
    }
}
