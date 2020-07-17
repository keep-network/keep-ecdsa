pragma solidity 0.5.17;

/// @title Ether Transfer Receiver.
/// @dev This contract is for testing purposes only.
contract TestEtherReceiver {
    bool shouldFail;

    /// @notice Rejects ether transfers sent to this contract if the shouldFail
    /// flag is set to true.
    function() external payable {
        require(!shouldFail, "Payment rejected");
    }

    function setShouldFail(bool _value) public {
        shouldFail = _value;
    }
}
