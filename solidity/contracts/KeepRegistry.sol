pragma solidity ^0.5.4;

/// @title Keep Registry
/// @notice Contract handling Keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract KeepRegistry {
    uint256 internal requestIDSeq = 1;
    uint256 internal groupIDSeq = 1;

    event BondedECDSAKeepRequested(uint256 requestID, uint256 groupID, uint32 groupSize, uint32 dishonestThreshold); 
 
    /// @notice Request a new bonded ECDSA Keep.
    /// @dev TODO: This is a stub function - needs to be implemented.
    /// @param _groupSize Number of members in the group.
    /// @param _dishonestThreshold Maximum number of dishonest group members.
    /// @return Unique request identifier.
    function requestBondedECDSAKeep(uint32 _groupSize, uint32 _dishonestThreshold) public payable returns (uint256 _requestID) {
        _requestID = requestIDSeq++;
        uint256 _groupID = groupIDSeq++;

        emit BondedECDSAKeepRequested(_requestID, _groupID, _groupSize, _dishonestThreshold);

        return _requestID;
    }
}
