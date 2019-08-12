pragma solidity ^0.5.4;

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeep {
    // Owner of the keep.
    address owner;
    // List of keep members' addresses.
    address[] internal members;
    // Minimum number of honest keep members required to produce a signature.
    uint256 honestThreshold;
    // Signer's ECDSA public key serialized to 64-bytes, where X and Y coordinates
    // are padded with zeros to 32-byte each.
    bytes publicKey;

    // Notification that a signer's public key was set for the keep.
    event PublicKeySet(
        bytes publicKey
    );

    // Notification that the keep was requested to sign a digest.
    event SignatureRequested(
        bytes32 digest
    );

    constructor(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold
    ) public {
        owner = _owner;
        members = _members;
        honestThreshold = _honestThreshold;
    }

    /// @notice Set a signer's public key for the keep.
    /// @dev Stub implementations.
    /// @param _publicKey Signer's public key.
    function setPublicKey(bytes memory _publicKey) public {
        // TODO: Validate if `msg.sender` is on `members` list.
        // TODO: Validate format: 32-bytes X + 32-bytes Y.
        publicKey = _publicKey;
        emit PublicKeySet(_publicKey);
    }

    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() public view returns (bytes memory) {
       return publicKey;
    }

    /// @notice Calculates a signature over provided digest by the keep.
    /// @dev TODO: Access control.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) public {
        require(msg.sender == owner, "Only keep owner can ask to sign");
        emit SignatureRequested(_digest);
    }
}
