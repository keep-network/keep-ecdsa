pragma solidity ^0.5.4;

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
contract IBondedECDSAKeep {
    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() external view returns (bytes memory);

    /// @notice Returns the amount of the keep's ETH bond in wei.
    /// @return The amount of the keep's ETH bond in wei.
    function checkBondAmount() external view returns (uint256);

    /// @notice Calculates a signature over provided digest by the keep. Note that
    /// signatures from the keep not explicitly requested by calling `sign`
    /// will be provable as fraud via `submitSignatureFraud`.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) external;

    /// @notice Distributes ETH evenly across all keep members.
    /// @dev Only the value passed to this function will be distributed.
    function distributeETHToMembers() external payable;

    /// @notice Distributes ERC20 token evenly across all keep members.
    /// @dev This works with any ERC20 token that implements a transferFrom
    /// function.
    /// This function only has authority over pre-approved
    /// token amount. We don't explicitly check for allowance, SafeMath
    /// subtraction overflow is enough protection.
    /// @param _tokenAddress Address of the ERC20 token to distribute.
    /// @param _value Amount of ERC20 token to distribute.
    function distributeERC20ToMembers(address _tokenAddress, uint256 _value)
        external;

    /// @notice Seizes the signers' ETH bond. After seizing bonds keep is
    /// closed so it will not respond to signing requests. Bonds can be seized
    /// only when there is no signing in progress or requested signing process
    /// has timed out.
    function seizeSignerBonds() external;

    /// @notice Submits a fraud proof for a valid signature from this keep that was
    /// not first approved via a call to sign.
    /// @dev The function expects the signed digest to be calculated as a double
    /// sha256 hash (hash256) of the preimage: `sha256(sha256(_preimage))`.
    /// @param _v Signature's header byte: `27 + recoveryID`.
    /// @param _r R part of ECDSA signature.
    /// @param _s S part of ECDSA signature.
    /// @param _signedDigest Digest for the provided signature. Result of hashing
    /// the preimage.
    /// @param _preimage Preimage of the hashed message.
    /// @return True if fraud, error otherwise.
    function submitSignatureFraud(
        uint8 _v,
        bytes32 _r,
        bytes32 _s,
        bytes32 _signedDigest,
        bytes calldata _preimage
    ) external returns (bool _isFraud);

    /// @notice Closes keep when owner decides that they no longer need it.
    /// Releases bonds to the keep members. Keep can be closed only when
    /// there is no signing in progress or requested signing process has timed out.
    function closeKeep() external;
}
