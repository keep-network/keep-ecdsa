pragma solidity 0.5.17;

contract TokenGrantStub {
    mapping(address => address[]) granteeOperators;

    function getGranteeOperators(
        address grantee
    ) public view returns(address[] memory) {
        return granteeOperators[grantee];
    }

    function setGranteeOperator(address grantee, address operator) public {
        address[] memory operators = new address[](1);
        operators[0] = operator;
        granteeOperators[grantee] = operators;
    }
}