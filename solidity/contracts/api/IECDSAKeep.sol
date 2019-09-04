pragma solidity ^0.5.4;

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract IECDSAKeep {

    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() public view returns (bytes memory);

    /// @notice Calculates a signature over provided digest by the keep.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) public;
}
