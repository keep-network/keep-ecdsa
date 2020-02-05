pragma solidity ^0.5.4;

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract IBondedECDSAKeep {

    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() external view returns (bytes memory);

    /// @notice Returns the amount of the keep's ETH bond in wei.
    /// @return The amount of the keep's ETH bond in wei.
    function checkBondAmount() external returns (uint256);

    /// @notice Calculates a signature over provided digest by the keep. Note that
    ///         signatures from the keep not explicitly requested by calling `sign`
    ///         will be provable as fraud via `submitSignatureFraud`.
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
    function distributeERC20ToMembers(address _tokenAddress, uint256 _value) external;

    /// @notice Seizes the signer's ETH bond.
    function seizeSignerBonds() external;

    /// @notice Submits a fraud proof for a valid signature from this keep that was
    ///         not first approved via a call to sign.
    /// @return Error if not fraud, true if fraud.
    function submitSignatureFraud(
        uint8 _v,
        bytes32 _r,
        bytes32 _s,
        bytes32 _signedDigest,
        bytes calldata _preimage
    ) external returns (bool _isFraud);
}
