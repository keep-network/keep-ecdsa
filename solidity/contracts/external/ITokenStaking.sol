pragma solidity ^0.5.4;

/// @title Token Staking interface.
/// @notice Interface for interaction with TokenStaking contract.
contract ITokenStaking {
    ///@dev Gets the eligible stake balance of the specified address.
    /// An eligible stake is a stake that passed the initialization period
    /// and is not currently undelegating. Also, the operator had to approve
    /// the specified operator contract.
    ///
    /// Operator with a minimum required amount of eligible stake can join the
    /// network and participate in new work selection.
    ///
    /// @param _operator address of stake operator.
    /// @param _operatorContract address of operator contract.
    /// @return an uint256 representing the eligible stake balance.
    function eligibleStake(address _operator, address _operatorContract)
        public
        view
        returns (uint256 balance);
}
