pragma solidity ^0.5.4;

import "./ECDSAKeepFactory.sol";

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeep {
    // Factory which built the keep.
    ECDSAKeepFactory factory;
    // Owner of the keep.
    address owner;
    // List of keep members' unique identifiers.
    uint256[] internal keepMembers;
    // Maximum number of dishonest members in the keep.
    uint256 dishonestThreshold;

    constructor(
        address _owner,
        uint256[] memory _keepMembers,
        uint256 _dishonestThreshold
    ) public {
        factory = ECDSAKeepFactory(msg.sender);
        owner = _owner;
        keepMembers = _keepMembers;
        dishonestThreshold = _dishonestThreshold;
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
