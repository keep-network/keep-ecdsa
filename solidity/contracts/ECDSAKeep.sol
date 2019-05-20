pragma solidity ^0.5.4;

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeep {
    // Owner of the keep.
    address owner;
    // List of keep members' addresses.
    address[] internal members;
    // Minimum number of honest members in the keep.
    uint256 honestThreshold;

    constructor(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold
    ) public {
        owner = _owner;
        members = _members;
        honestThreshold = _honestThreshold;
    }

    /// @notice Calculates a signature over provided digest by the keep.
    /// @dev Stub implementations it should be calling sMPC cluster to produce
    /// a signature.
    /// @param _digest Digest to be signed.
    function sign(bytes memory _digest) public view {
        require(msg.sender == owner, "Only keep owner can ask to sign");
        // TODO: Emit event to sign the digest.
        _digest;
    }
}
