pragma solidity ^0.5.4;

/// @title Keep Bonding
/// @notice Contract holding deposits from keeps' operators.
contract KeepBonding {
    // Unassigned ether values deposited by operators.
   mapping(address => uint256) internal pot;
   // References to created bonds.
   mapping(bytes => uint256) internal lockedBonds;
   mapping(bytes => uint256[]) internal bondAssignments;

   /// @notice Returns value of ether available for bonding for the operator.
   /// @param operator Address of the operator.
   /// @return Value of deposited ether available for bonding.
   function availableBondingValue(address operator) public view returns (uint256) {
      return pot[operator];
   }

   /// @notice Add ether to sender's value available for bonding.
   function deposit(address operator) external payable {
      pot[operator] += msg.value;
   }

   /// @notice Draw amount from sender's value available for bonding.
   /// @param amount Value to withdraw.
   /// @param destination Address to send amount.
   function withdraw(uint256 amount, address payable destination) external {
      require(availableBondingValue(msg.sender) >= amount, "Insufficient pot");

      pot[msg.sender] -= amount;
      destination.transfer(amount);
   }

   /// @notice Create bond for given operator, reference and amount.
   /// @dev Function can be executed only by authorized contract which will become
   /// bond's holder.
   /// @param operator Address of the operator to bond.
   /// @param ref Reference used to track the bond by holder.
   /// @param amount Value to bond.
   function createBond(address operator, uint256 ref, uint256 amount) public onlyAuthorized {
      require(availableBondingValue(operator) >= amount, "Insufficient pot");

      address holder = msg.sender;
      bytes memory bondID = abi.encodePacked(operator, holder, ref);
      bytes memory assignment = abi.encodePacked(operator, holder);

      pot[operator] -= amount;
      lockedBonds[bondID] += amount;
      bondAssignments[assignment].push(ref);
   }

   /// @notice Checks if the caller is an authorized contract.
   /// @dev Throws an error if called by any account other than one of the authorized
   /// contracts.
   modifier onlyAuthorized() {
      // TODO: Add authorization checks.
      _;
   }
}
