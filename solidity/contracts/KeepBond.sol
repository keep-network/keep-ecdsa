pragma solidity ^0.5.4;

/// @title Keep Bond
/// @notice Contract holding deposits from keeps' operators.
contract KeepBond {
   // Unassigned ether values deposited by operators.
   mapping(address => uint) internal pot;

   /// @notice Add ether to sender's value available for bonding.
   function deposit() external payable {
      pot[msg.sender] += msg.value;
   }

   /// @notice Draw amount from sender's value available for bonding.
   /// @param amount Value to withdraw.
   /// @param destination Address to send amount.
   function withdraw(uint amount, address payable destination) external {
      require(pot[msg.sender] >= amount, "Insufficient pot");

      pot[msg.sender] -= amount;
      destination.transfer(amount);
   }
}
