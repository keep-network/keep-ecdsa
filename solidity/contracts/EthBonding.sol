pragma solidity ^0.5.4;

import "./AbstractBonding.sol";

import "@keep-network/keep-core/contracts/Registry.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

// TODO: This contract is expected to implement functions defined by IBonding
// interface defined in @keep-network/sortition-pools. After merging the
// repositories we need to move IBonding definition to sit closer to KeepBonding
// contract so that sortition pools import it for own needs. It is the bonding
// module which should define an interface, and sortition pool module should be
// just importing it.

/// @title Eth Bonding
/// @notice Contract holding deposits from keeps' operators.
contract EthBonding is AbstractBonding {
    using SafeMath for uint256;

    // Registry contract with a list of approved factories (operator contracts).
    Registry internal registry;

    uint256 internal initializationPeriod;

    mapping(address => uint256) public createdAt;

    // Operator contracts authorized by operators.
    mapping(address => mapping(address => bool)) internal authorizedOperatorContracts;

    /// @notice Initializes Keep Bonding contract.
    /// @param registryAddress Keep registry contract address.
    /// @param _initializationPeriod Initialization period.
    constructor(address registryAddress, uint256 _initializationPeriod) public {
        registry = Registry(registryAddress);
        initializationPeriod = _initializationPeriod;
    }

    function isInitialized(address operator) public view returns (bool) {
        return block.number > createdAt[operator].add(initializationPeriod);
    }

    /// @notice Returns the amount of wei the operator has made available for
    /// bonding and that is still unbounded. If the operator doesn't exist or
    /// bond creator is not authorized as an operator contract or it is not
    /// authorized by the operator or there is no secondary authorization for
    /// the provided sortition pool, function returns 0.
    /// @dev Implements function expected by sortition pools' IBonding interface.
    /// @param operator Address of the operator.
    /// @param bondCreator Address authorized to create a bond.
    /// @param authorizedSortitionPool Address of authorized sortition pool.
    /// @return Amount of authorized wei deposit available for bonding.
    function availableUnbondedValue(
        address operator,
        address bondCreator,
        address authorizedSortitionPool
    ) public view returns (uint256) {
        // Sortition pools check this condition and skips operators that
        // are no longer eligible. We cannot revert here.
        if (
            isInitialized(operator) &&
            registry.isApprovedOperatorContract(bondCreator) &&
            authorizedOperatorContracts[operator][bondCreator] &&
            hasSecondaryAuthorization(operator, authorizedSortitionPool)
        ) {
          return unbondedValue[operator];
        }

        return 0;
    }

    /// @notice Add the provided value to operator's pool available for bonding.
    /// @param operator Address of the operator.
    function deposit(address operator) external payable {
        require(
            createdAt[operator] == 0,
            "Operator already in use"
        );
        unbondedValue[operator] = msg.value;
        createdAt[operator] = block.number;
    }

    function authorizeOperatorContract(
        address _operator,
        address _contract
    ) public {
        require(
            _isAuthorizerOf(msg.sender, _operator),
            "May not authorize for other addresses"
        );
        authorizedOperatorContracts[_operator][_contract] = true;
    }

    function _isWithdrawerOf(
        address _sender,
        address _operator
    ) internal view returns (bool) {
        return _sender == _operator;
    }

    function _isAuthorizerOf(
        address _sender,
        address _operator
    ) internal view returns (bool) {
        return _sender == _operator;
    }

    function _beneficiaryOf(
        address _operator
    ) internal view returns (address payable) {
        return address(uint160(_operator));
    }
}
