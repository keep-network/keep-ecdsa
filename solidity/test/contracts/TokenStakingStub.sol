pragma solidity ^0.5.4;
import "@keep-network/sortition-pools/contracts/api/IStaking.sol";

/// @title Token Staking Stub
/// @dev This contract is for testing purposes only.
contract TokenStakingStub is IStaking {
    uint256 balance;

    mapping(address => address payable) operatorToMagpie;

    // Authorized operator contracts.
    mapping(address => mapping (address => bool)) internal authorizations;

    // Map of operator -> owner.
    mapping(address => address) owners;

    /// @dev Sets balance variable value.
    function setBalance(uint256 _balance) public {
        balance = _balance;
    }

    /// @dev Returns balance variable value.
    function eligibleStake(address _address, address)
        public
        view
        returns (uint256)
    {
        return balance;
    }

    function setMagpie(address _operator, address payable _magpie) public {
        operatorToMagpie[_operator] = _magpie;
    }

    function magpieOf(address _operator) public view returns (address payable) {
        address payable magpie = operatorToMagpie[_operator];
        if (magpie == address(0)) {
            return address(uint160(_operator));
        }
        return magpie;
    }

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
        return _operator;
    }

    function setOwner(address _operator, address _owner) public {
        owners[_operator] = _owner;
    }

    function ownerOf(address _operator) public view returns (address) {
        return owners[_operator];
    }
}
