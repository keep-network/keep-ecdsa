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

/// @title ETH Bonding
/// @notice Contract holding deposits and delegations for ETH-only keeps'
/// operators. An owner of the ETH can delegate ETH to an operator. The value
/// of ETH the owner is willing to delegate should be deposited for the given
/// operator.
contract EthBonding is AbstractBonding, Authorizations, StakeDelegatable {
    event Delegated(address indexed owner, address indexed operator);

    event OperatorDelegated(
        address indexed operator,
        address indexed beneficiary,
        address indexed authorizer
    );

    /// @notice Initializes Keep Bonding contract.
    /// @param _keepRegistry Keep Registry contract address.
    constructor(KeepRegistry _keepRegistry)
        public
        AbstractBonding(_keepRegistry)
        Authorizations(_keepRegistry)
    {}

    /// @notice Registers delegation details. The function is used to register
    /// addresses of operator, beneficiary and authorizer for a delegation from
    /// the caller.
    /// @param operator Address of the operator.
    /// @param beneficiary Address of the beneficiary.
    /// @param authorizer Address of the authorizer.
    function delegate(
        address operator,
        address payable beneficiary,
        address authorizer
    ) external payable {
        // check value passed with delegation

        address owner = msg.sender;

        require(
            operators[operator].owner == address(0),
            "Operator already in use"
        );

        operators[operator] = Operator(0, owner, beneficiary, authorizer);

        // TODO: Add initialization period as per https://github.com/keep-network/keep-ecdsa/pull/483#discussion_r468628872

        emit Delegated(owner, operator);
        emit OperatorDelegated(operator, beneficiary, authorizer);
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
}
