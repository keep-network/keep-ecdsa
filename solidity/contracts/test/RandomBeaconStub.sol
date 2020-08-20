pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/IRandomBeacon.sol";

/// @title Random Beacon Service Stub
/// @dev This contract is for testing purposes only.
contract RandomBeaconStub is IRandomBeacon {
    uint256 feeEstimate = 58;
    uint256 entry = 0;
    uint256 public requestCount = 0;
    bool shouldFail;

    function requestRelayEntry() external payable returns (uint256) {
        return requestRelayEntry(address(0), 0);
    }

    /// @dev Get the entry fee estimate in wei for relay entry request.
    /// @param callbackGas Gas required for the callback.
    function entryFeeEstimate(uint256 callbackGas)
        public
        view
        returns (uint256)
    {
        return feeEstimate;
    }

    function requestRelayEntry(
        address callbackContract,
        uint256 callbackGas
    ) public payable returns (uint256) {
        requestCount++;

        if (shouldFail) {
            revert("request relay entry failed");
        }

        if (entry != 0) {
            callbackContract.call(
                abi.encodeWithSignature("__beaconCallback(uint256)", entry)
            );
        }

        return requestCount;
    }

    function setEntry(uint256 newEntry) public {
        entry = newEntry;
    }

    function setShouldFail(bool value) public {
        shouldFail = value;
    }
}
