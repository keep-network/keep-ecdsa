pragma solidity ^0.5.4;

/// @title Keep Registry
/// @notice Contract handling Keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract KeepRegistry {
    uint256 internal requestIDSeq = 1;
    uint256 internal groupIDSeq = 1;

    // TODO: We don't need to emit this event. Selection will be performed 
    // on-chain and we will emit an event when Keep is formed.
    event BondedECDSAKeepRequested(uint256 requestID, uint256 groupID, uint32 groupSize, uint32 dishonestThreshold); 
 
    /// @notice Request a new bonded ECDSA Keep.
    /// @dev TODO: This is a stub function - needs to be implemented.
    /// @param _groupSize Number of members in the group.
    /// @param _dishonestThreshold Maximum number of dishonest group members.
    /// @param _bond Required bond (in Gwei) from each member.
    /// @return Unique request identifier.
    function requestBondedECDSAKeep(uint32 _groupSize, uint32 _dishonestThreshold, uint256 _bond) public payable returns (uint256 _requestID) {
        _requestID = requestIDSeq++;
        uint256 _groupID = groupIDSeq++;
        _bond;

        // TODO: Select group on-chain and emit an event that a group was registered
        // with specific members.
        emit BondedECDSAKeepRequested(_requestID, _groupID, _groupSize, _dishonestThreshold);

        return _requestID;
    }
}
