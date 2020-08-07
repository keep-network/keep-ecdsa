pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/Authorizations.sol";
import "@keep-network/keep-core/contracts/KeepRegistry.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Authorizations Stub
/// @dev This contract is for testing purposes only.
contract AuthorizationsStub is Authorizations {
    // Authorized operator contracts.
    mapping(address => mapping(address => bool)) internal authorizations;

    address public delegatedAuthority;

    constructor(KeepRegistry _registry) public Authorizations(_registry) {}

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

    function authorizerOf(address _operator) public view returns (address) {
        revert("abstract function");
    }

    function claimDelegatedAuthority(address delegatedAuthoritySource) public {
        delegatedAuthority = delegatedAuthoritySource;
    }
}
