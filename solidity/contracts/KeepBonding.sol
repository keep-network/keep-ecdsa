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

import "@keep-network/keep-core/contracts/TokenGrant.sol";
import "@keep-network/keep-core/contracts/libraries/RolesLookup.sol";


/// @title Keep Bonding
/// @notice Contract holding deposits from keeps' operators.
contract KeepBonding is AbstractBonding {
    using RolesLookup for address payable;

    // KEEP token grant contract.
    TokenGrant internal tokenGrant;

    /// @notice Initializes Keep Bonding contract.
    /// @param registryAddress Keep registry contract address.
    /// @param tokenStakingAddress KEEP token staking contract address.
    /// @param tokenGrantAddress KEEP token grant contract address.
    constructor(
        address registryAddress,
        address tokenStakingAddress,
        address tokenGrantAddress
    )
        public
        AbstractBonding(
            registryAddress,
            tokenStakingAddress, // Authorizations
            tokenStakingAddress // StakeDelegatable
        )
    {
        tokenGrant = TokenGrant(tokenGrantAddress);
    }

    /// @notice Withdraws amount from operator's value available for bonding.
    /// Should not be used by grantee of managed grants. For this case,
    /// please use `withdrawAsManagedGrantee`.
    ///
    /// This function can be called only by:
    /// - operator,
    /// - liquid, staked tokens owner (not a grant),
    /// - direct staked tokens grantee (not a managed grant).
    ///
    /// @param amount Value to withdraw in wei.
    /// @param operator Address of the operator.
    function withdraw(uint256 amount, address operator) public {
        require(
            msg.sender == operator ||
                msg.sender.isTokenOwnerForOperator(
                    operator,
                    stakeDelegatable
                ) ||
                msg.sender.isGranteeForOperator(operator, tokenGrant),
            "Only operator or the owner is allowed to withdraw bond"
        );

        withdrawBond(amount, operator);
    }

    /// @notice Withdraws amount from operator's value available for bonding.
    /// Can be called only by staked tokens managed grantee.
    /// @param amount Value to withdraw in wei.
    /// @param operator Address of the operator.
    /// @param managedGrant Address of the managed grant contract.
    function withdrawAsManagedGrantee(
        uint256 amount,
        address operator,
        address managedGrant
    ) public {
        require(
            msg.sender.isManagedGranteeForOperator(
                operator,
                managedGrant,
                tokenGrant
            ),
            "Only grantee is allowed to withdraw bond"
        );

        withdrawBond(amount, operator);
    }
}
