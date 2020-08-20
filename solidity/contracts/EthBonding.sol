/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓         ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "./AbstractBonding.sol";

import "@keep-network/keep-core/contracts/Authorizations.sol";
import "@keep-network/keep-core/contracts/KeepRegistry.sol";
import "@keep-network/keep-core/contracts/StakeDelegatable.sol";

import "@keep-network/sortition-pools/contracts/api/IFullyBackedBonding.sol";

/// @title ETH Bonding
/// @notice Contract holding deposits and delegations for ETH-only keeps'
/// operators. An owner of the ETH can delegate ETH to an operator. The value
/// of ETH the owner is willing to delegate should be deposited for the given
/// operator.
contract EthBonding is
    AbstractBonding,
    Authorizations,
    StakeDelegatable,
    IFullyBackedBonding
{
    event Delegated(address indexed owner, address indexed operator);

    event OperatorDelegated(
        address indexed operator,
        address indexed beneficiary,
        address indexed authorizer
    );

    // TODO: Decide on the final value and if we want a setter for it.
    uint256 public constant MINIMUM_DELEGATION_DEPOSIT = 12345;

    uint256 public initializationPeriod;

    /// @notice Initializes Keep Bonding contract.
    /// @param _keepRegistry Keep Registry contract address.
    /// @param _initializationPeriod To avoid certain attacks on group selection,
    /// recently delegated operators must wait for a specific period of time
    /// before being eligible for group selection.
    constructor(KeepRegistry _keepRegistry, uint256 _initializationPeriod)
        public
        AbstractBonding(_keepRegistry)
        Authorizations(_keepRegistry)
    {
        initializationPeriod = _initializationPeriod;
    }

    /// @notice Registers delegation details. The function is used to register
    /// addresses of operator, beneficiary and authorizer for a delegation from
    /// the caller.
    /// The function requires ETH to be submitted in the call as a protection
    /// against attacks blocking operators. The value should be at least equal
    /// to the minimum delegation deposit. Whole amount is deposited as operator's
    /// unbonded value for the future bonding.
    /// @param operator Address of the operator.
    /// @param beneficiary Address of the beneficiary.
    /// @param authorizer Address of the authorizer.
    function delegate(
        address operator,
        address payable beneficiary,
        address authorizer
    ) external payable {
        address owner = msg.sender;

        require(
            operators[operator].owner == address(0),
            "Operator already in use"
        );

        require(
            msg.value >= MINIMUM_DELEGATION_DEPOSIT,
            "Insufficient delegation value"
        );

        operators[operator] = Operator(
            OperatorParams.pack(0, block.timestamp, 0),
            owner,
            beneficiary,
            authorizer
        );

        deposit(operator);

        emit Delegated(owner, operator);
        emit OperatorDelegated(operator, beneficiary, authorizer);
    }

    /// @notice Checks if the operator for the given bond creator contract
    /// has passed the initialization period.
    /// @param operator The operator address.
    /// @param bondCreator The bond creator contract address.
    /// @return True if operator has passed initialization period for given
    /// bond creator contract, false otherwise.
    function isInitialized(address operator, address bondCreator)
        public
        view
        returns (bool)
    {
        uint256 operatorParams = operators[operator].packedParams;

        return
            isAuthorizedForOperator(operator, bondCreator) &&
            _isInitialized(operatorParams);
    }

    /// @notice Withdraws amount from operator's value available for bonding.
    /// This function can be called only by:
    /// - operator,
    /// - owner of the stake.
    ///
    /// @param amount Value to withdraw in wei.
    /// @param operator Address of the operator.
    function withdraw(uint256 amount, address operator) public {
        require(
            msg.sender == operator || msg.sender == ownerOf(operator),
            "Only operator or the owner is allowed to withdraw bond"
        );

        withdrawBond(amount, operator);
    }

    /// @notice Is the operator with the given params initialized
    function _isInitialized(uint256 operatorParams)
        internal
        view
        returns (bool)
    {
        return
            block.timestamp >
            operatorParams.getCreationTimestamp().add(initializationPeriod);
    }
}
