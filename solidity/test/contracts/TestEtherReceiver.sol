pragma solidity ^0.5.4;

/// @title Ether Transfer Receiver.
/// @dev This contract is for testing purposes only.
contract TestEtherReceiver {
    uint256 public invalidValue = 666;

    /// @notice Rejects ether transfers sent to this contract if the value equals
    /// `invalidValue`.
    function() external payable {
        require(msg.value != invalidValue, "Payment rejected");
    }
}
