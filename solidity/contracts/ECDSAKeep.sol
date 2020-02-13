pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "@keep-network/keep-core/contracts/TokenStaking.sol";
import "@keep-network/keep-core/contracts/utils/AddressArrayUtils.sol";
import "./api/IBondedECDSAKeep.sol";
import "./KeepBonding.sol";

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
contract ECDSAKeep is IBondedECDSAKeep, Ownable {
    using AddressArrayUtils for address[];
    using SafeMath for uint256;

    // List of keep members' addresses.
    address[] internal members;

    // Minimum number of honest keep members required to produce a signature.
    uint256 honestThreshold;

    // Signer's ECDSA public key serialized to 64-bytes, where X and Y coordinates
    // are padded with zeros to 32-byte each.
    bytes publicKey;

    // Latest digest requested to be signed. Used to validate submitted signature.
    bytes32 digest;

    // Map of all digests requested to be signed. Used to validate submitted signature.
    mapping(bytes32 => bool) digests;

    // Timeout in blocks for a signature to appear on the chain. Blocks are
    // counted from the moment signing request occurred.
    uint256 public signingTimeout = 90 * 60;  // [seconds]

    // The timestamp at which signing process started. Used also to track if
    // signing is in progress. When set to `0` indicates there is no
    // signing process in progress.
    uint256 internal signingStartTimestamp;

    // Notification that a signer's public key was published for the keep.
    event PublicKeyPublished(bytes publicKey);

    // Notification that the keep was requested to sign a digest.
    event SignatureRequested(bytes32 digest);

    // Notification that the signature has been calculated. Contains a digest which
    // was used for signature calculation and a signature in a form of r, s and
    // recovery ID values.
    // The signature is chain-agnostic. Some chains (e.g. Ethereum and BTC) requires
    // `v` to be calculated by increasing recovery id by 27. Please consult the
    // documentation about what the particular chain expects.
    event SignatureSubmitted(
        bytes32 digest,
        bytes32 r,
        bytes32 s,
        uint8 recoveryID
    );

    KeepBonding keepBonding;

    TokenStaking tokenStaking;

    constructor(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold,
        address _keepBonding,
        address _tokenStaking
    ) public {
        transferOwnership(_owner);
        members = _members;
        honestThreshold = _honestThreshold;
        keepBonding = KeepBonding(_keepBonding);
        tokenStaking = TokenStaking(_tokenStaking);
    }

    /// @notice Set a signer's public key for the keep.
    /// @dev Stub implementations.
    /// @param _publicKey Signer's public key.
    function setPublicKey(bytes calldata _publicKey) external onlyMember {
        require(_publicKey.length == 64, "Public key must be 64 bytes long");
        publicKey = _publicKey;
        emit PublicKeyPublished(_publicKey);
    }

    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() external view returns (bytes memory) {
        return publicKey;
    }

    /// @notice Returns the amount of the keep's ETH bond in wei.
    /// @return The amount of the keep's ETH bond in wei.
    function checkBondAmount() external view returns (uint256) {
        uint256 sumBondAmount = 0;
        for (uint256 i = 0; i < members.length; i++) {
            sumBondAmount += keepBonding.bondAmount(
                members[i],
                address(this),
                uint256(address(this))
            );
        }

        return sumBondAmount;
    }

    /// @notice Seizes the signer's ETH bond.
    function seizeSignerBonds() external onlyOwner {
        for (uint256 i = 0; i < members.length; i++) {
            uint256 amount = keepBonding.bondAmount(
                members[i],
                address(this),
                uint256(address(this))
            );

            keepBonding.seizeBond(
                members[i],
                uint256(address(this)),
                amount,
                address(uint160(owner()))
            );
        }
    }

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
    ) external returns (bool _isFraud) {
        bytes32 calculatedDigest = sha256(abi.encodePacked(sha256(_preimage)));
        require(
            _signedDigest == calculatedDigest,
            "Signed digest does not match double sha256 hash of the preimage"
        );

        bool isSignatureValid = publicKeyToAddress(publicKey) ==
            ecrecover(_signedDigest, _v, _r, _s);

        // Check if the signature is valid but was not requested.
        require(
            isSignatureValid && !digests[_signedDigest],
            "Signature is not fraudulent"
        );

        return true;
    }

    /// @notice Calculates a signature over provided digest by the keep.
    /// @dev Only one signing process can be in progress at a time.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) external onlyOwner {
        require(!isSigningInProgress(), "Signer is busy");

        /* solium-disable-next-line */
        signingStartTimestamp = block.timestamp;

        digests[_digest] = true;
        digest = _digest;

        emit SignatureRequested(_digest);
    }

    /// @notice Submits a signature calculated for the given digest.
    /// @dev Fails if signature has not been requested or a signature has already
    /// been submitted.
    /// Validates s value to ensure it's in the lower half of the secp256k1 curve's
    /// order.
    /// @param _r Calculated signature's R value.
    /// @param _s Calculated signature's S value.
    /// @param _recoveryID Calculated signature's recovery ID (one of {0, 1, 2, 3}).
    function submitSignature(bytes32 _r, bytes32 _s, uint8 _recoveryID)
        external
        onlyMember
    {
        require(isSigningInProgress(), "Not awaiting a signature");
        require(!hasSigningTimedOut(), "Signing timeout elapsed");
        require(_recoveryID < 4, "Recovery ID must be one of {0, 1, 2, 3}");

        // Validate `s` value for a malleability concern described in EIP-2.
        // Only signatures with `s` value in the lower half of the secp256k1
        // curve's order are considered valid.
        require(
            uint256(_s) <=
                0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0,
            "Malleable signature - s should be in the low half of secp256k1 curve's order"
        );

        // We add 27 to the recovery ID to align it with ethereum and bitcoin
        // protocols where 27 is added to recovery ID to indicate usage of
        // uncompressed public keys.
        uint8 _v = 27 + _recoveryID;

        // Validate signature.
        require(
            publicKeyToAddress(publicKey) == ecrecover(digest, _v, _r, _s),
            "Invalid signature"
        );

        signingStartTimestamp = 0;

        emit SignatureSubmitted(digest, _r, _s, _recoveryID);
    }

    /// @notice Returns true if signing of a digest is currently in progress.
    function isSigningInProgress() internal view returns (bool) {
        return signingStartTimestamp != 0;
    }

    /// @notice Returns true if the ongoing signing process timed out.
    /// @dev There is a certain timeout for a signature to be produced, see
    /// `signingTimeout`.
    function hasSigningTimedOut() internal view returns (bool) {
        return
            signingStartTimestamp != 0 &&
            /* solium-disable-next-line */
            block.timestamp > signingStartTimestamp + signingTimeout;
    }



    /// @notice Coverts a public key to an ethereum address.
    /// @param _publicKey Public key provided as 64-bytes concatenation of
    /// X and Y coordinates (32-bytes each).
    /// @return Ethereum address.
    function publicKeyToAddress(bytes memory _publicKey)
        internal
        pure
        returns (address)
    {
        // We hash the public key and then truncate last 20 bytes of the digest
        // which is the ethereum address.
        return address(uint160(uint256(keccak256(_publicKey))));
    }

    /// @notice Distributes ETH evenly across all keep members.
    /// ETH is sent to the beneficiary of each member.
    /// @dev Only the value passed to this function will be distributed.
    function distributeETHToMembers() external payable {
        uint256 memberCount = members.length;
        uint256 dividend = msg.value.div(memberCount);

        require(dividend > 0, "dividend value must be non-zero");

        for (uint16 i = 0; i < memberCount; i++) {
            // We don't want to revert the whole execution in case of single
            // transfer failure, hence we don't validate it's result.
            // TODO: What should we do with the dividend which was not transferred
            // successfully?
            tokenStaking.magpieOf(members[i]).call.value(dividend)("");
        }
    }

    /// @notice Distributes ERC20 token evenly across all keep members.
    /// The token is sent to the beneficiary of each member.
    /// @dev This works with any ERC20 token that implements a transferFrom
    /// function similar to the interface imported here from
    /// openZeppelin. This function only has authority over pre-approved
    /// token amount. We don't explicitly check for allowance, SafeMath
    /// subtraction overflow is enough protection.
    /// @param _tokenAddress Address of the ERC20 token to distribute.
    /// @param _value Amount of ERC20 token to distribute.
    function distributeERC20ToMembers(address _tokenAddress, uint256 _value)
        external
    {
        IERC20 token = IERC20(_tokenAddress);

        uint256 memberCount = members.length;
        uint256 dividend = _value.div(memberCount);

        require(dividend > 0, "dividend value must be non-zero");

        for (uint16 i = 0; i < memberCount; i++) {
            token.transferFrom(
                msg.sender,
                tokenStaking.magpieOf(members[i]),
                dividend
            );
        }
    }

    /// @notice Checks if the caller is a keep member.
    /// @dev Throws an error if called by any account other than one of the members.
    modifier onlyMember() {
        require(members.contains(msg.sender), "Caller is not the keep member");
        _;
    }

}
