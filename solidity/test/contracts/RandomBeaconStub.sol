pragma solidity ^0.5.4;

import {IRandomBeaconService} from "../../contracts/ECDSAKeepFactory.sol";

/// @title Random Beacon Service Stub
/// @dev This contract is for testing purposes only.
contract RandomBeaconStub is IRandomBeaconService {
    uint256 feeEstimate = 58;
    uint256 entry = 0;
    uint256 public calledTimes = 0;
    bool shouldFail;

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
        string memory callbackMethod,
        uint256 callbackGas
    ) public payable returns (uint256) {
        calledTimes++;

        if (shouldFail) {
            revert("request relay entry failed");
        }

        if (entry != 0) {
            callbackContract.call(
                abi.encodeWithSignature(callbackMethod, entry)
            );
        }
    }

    function setEntry(uint256 newEntry) public {
        entry = newEntry;
    }

    function setShouldFail(bool value) public {
        shouldFail = value;
    }
}
