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

import "../AbstractBonding.sol";

import "@keep-network/keep-core/contracts/Authorizations.sol";
import "@keep-network/keep-core/contracts/StakeDelegatable.sol";
import "@keep-network/keep-core/contracts/KeepRegistry.sol";

import "@keep-network/sortition-pools/contracts/api/IFullyBackedBonding.sol";

/// @title Fully Backed Bonding
/// @notice Contract holding deposits and delegations for ETH-only keeps'
/// operators. An owner of the ETH can delegate ETH to an operator by depositing
/// it in this contract.
contract FullyBackedBonding is
    IFullyBackedBonding,
    AbstractBonding,
    Authorizations,
    StakeDelegatable
{
    event Delegated(address indexed owner, address indexed operator);

    event OperatorDelegated(
        address indexed operator,
        address indexed beneficiary,
        address indexed authorizer,
        uint256 value
    );

    event OperatorToppedUp(address indexed operator, uint256 value);

    // The ether value (in wei) that should be passed along with the delegation
    // and deposited for bonding.
    uint256 public constant MINIMUM_DELEGATION_DEPOSIT = 40 ether;

    // Once a delegation to an operator is received the delegator has to wait for
    // specific time period before being able to pull out the funds.
    uint256 public constant DELEGATION_LOCK_PERIOD = 12 hours;

    uint256 public initializationPeriod; // varies between mainnet and testnet

    /// @notice Initializes Fully Backed Bonding contract.
    /// @param _keepRegistry Keep Registry contract address.
    /// @param _initializationPeriod To avoid certain attacks on group selection,
    /// recently delegated operators must wait for a specific period of time
    /// before being eligible for group selection.
    constructor(KeepRegistry _keepRegistry, uint256 _initializationPeriod)
        public
        AbstractBonding(address(_keepRegistry))
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

        require(operator != address(0), "Invalid operator address");
        require(authorizer != address(0), "Invalid authorizer address");

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
        emit OperatorDelegated(operator, beneficiary, authorizer, msg.value);
    }

    /// @notice Top-ups operator's unbonded value.
    /// @dev This function should be used to add new unbonded value to the system
    /// for an operator. The `deposit` function defined in parent abstract contract
    /// should be called only by applications returning value that has been already
    /// initially deposited and seized later. As an application may seize bonds
    /// and return them to the bonding contract with `deposit` function it makes
    /// tracking the totally deposited value much more complicated. Functions
    /// `delegate` and `topUps` should be used to add fresh value to the contract
    /// and events emitted by these functions should be enough to determine total
    /// value deposited ever for an operator.
    /// @param operator Address of the operator.
    function topUp(address operator) public payable {
        deposit(operator);

        emit OperatorToppedUp(operator, msg.value);
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
    /// Withdraw cannot be performed immediately after delegation to protect
    /// from a griefing. It is required that delegator waits specific period
    /// of time before they can pull out the funds deposited on delegation.
    /// @param amount Value to withdraw in wei.
    /// @param operator Address of the operator.
    function withdraw(uint256 amount, address operator) public {
        require(
            msg.sender == operator || msg.sender == ownerOf(operator),
            "Only operator or the owner is allowed to withdraw bond"
        );

        require(
            hasDelegationLockPassed(operator),
            "Delegation lock period has not passed yet"
        );

        withdrawBond(amount, operator);
    }

    /// @notice Gets delegation info for the given operator.
    /// @param operator Address of the operator.
    /// @return createdAt The time when the delegation was created.
    /// @return undelegatedAt The time when undelegation has been requested.
    /// If undelegation has not been requested, 0 is returned.
    function getDelegationInfo(address operator)
        public
        view
        returns (uint256 createdAt, uint256 undelegatedAt)
    {
        uint256 operatorParams = operators[operator].packedParams;

        return (
            operatorParams.getCreationTimestamp(),
            operatorParams.getUndelegationTimestamp()
        );
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

    /// @notice Has lock period passed for a delegation.
    /// @param operator Address of the operator.
    /// @return True if delegation lock period passed, false otherwise.
    function hasDelegationLockPassed(address operator)
        internal
        view
        returns (bool)
    {
        return
            block.timestamp >
            operators[operator].packedParams.getCreationTimestamp().add(
                DELEGATION_LOCK_PERIOD
            );
    }
}
