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

    // Timeout for the keep public key to appear on the chain. Time is counted
    // from the moment keep has been created.
    uint256 public keyGenerationTimeout = 150 * 60; // 2.5h in seconds

    // The timestamp at which keep has been created and key generation process
    // started.
    uint256 keyGenerationStartTimestamp;

    // Timeout for a signature to appear on the chain. Time is counted from the
    // moment signing request occurred.
    uint256 public signingTimeout = 90 * 60; // 1.5h in seconds

    // The timestamp at which signing process started. Used also to track if
    // signing is in progress. When set to `0` indicates there is no
    // signing process in progress.
    uint256 internal signingStartTimestamp;

    // Map stores public key by member addresses. All members should submit the
    // same public key.
    mapping(address => bytes) submittedPublicKeys;

    // Notification that a signer's public key was published for the keep.
    event PublicKeyPublished(bytes publicKey);

    // Flag to monitor current keep state. If the keep is active members monitor
    // it and support requests for the keep owner. If the owner decides to close
    // the keep the flag is set to false.
    bool internal isActive;

    // Notification that the keep was requested to sign a digest.
    event SignatureRequested(bytes32 digest);

    // Notification that the submitted public key does not match a key submitted
    // by other member. The event contains address of the member who tried to
    // submit a public key and a conflicting public key submitted already by other
    // member.
    event ConflictingPublicKeySubmitted(
        address submittingMember,
        bytes conflictingPublicKey
    );

    // Notification that the keep was closed by the owner. Members no longer need
    // to support it.
    event KeepClosed();

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

    TokenStaking tokenStaking;
    KeepBonding keepBonding;

    constructor(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold,
        address _tokenStaking,
        address _keepBonding
    ) public {
        transferOwnership(_owner);
        members = _members;
        honestThreshold = _honestThreshold;
        tokenStaking = TokenStaking(_tokenStaking);
        keepBonding = KeepBonding(_keepBonding);
        isActive = true;

        /* solium-disable-next-line */
        keyGenerationStartTimestamp = block.timestamp;
    }

    /// @notice Submits a public key to the keep.
    /// @dev Public key is published successfully if all members submit the same
    /// value. In case of conflicts with others members submissions it will emit
    /// an `ConflictingPublicKeySubmitted` event. When all submitted keys match
    /// it will store the key as keep's public key and emit a `PublicKeyPublished`
    /// event.
    /// @param _publicKey Signer's public key.
    function submitPublicKey(bytes calldata _publicKey) external onlyMember {
        require(!hasKeyGenerationTimedOut(), "Key generation timeout elapsed");

        require(
            !hasMemberSubmittedPublicKey(msg.sender),
            "Member already submitted a public key"
        );

        require(_publicKey.length == 64, "Public key must be 64 bytes long");

        submittedPublicKeys[msg.sender] = _publicKey;

        // Check if public keys submitted by all keep members are the same as
        // the currently submitted one.
        uint256 matchingPublicKeysCount = 0;
        for (uint256 i = 0; i < members.length; i++) {
            if (
                keccak256(submittedPublicKeys[members[i]]) !=
                keccak256(_publicKey)
            ) {
                // Emit an event only if compared member already submitted a value.
                if (hasMemberSubmittedPublicKey(members[i])) {
                    emit ConflictingPublicKeySubmitted(
                        msg.sender,
                        submittedPublicKeys[members[i]]
                    );
                }
            } else {
                matchingPublicKeysCount++;
            }
        }

        if (matchingPublicKeysCount != members.length) {
            return;
        }

        // All submitted signatures match.
        publicKey = _publicKey;
        emit PublicKeyPublished(_publicKey);
    }

    /// @notice Returns true if the ongoing key generation process timed out.
    /// @dev There is a certain timeout for keep public key to be produced and
    /// appear on the chain, see `keyGenerationTimeout`.
    function hasKeyGenerationTimedOut() internal view returns (bool) {
        /* solium-disable-next-line */
        return block.timestamp > keyGenerationStartTimestamp + keyGenerationTimeout;
    }

    /// @notice Checks if the member already submitted a public key.
    /// @param _member Address of the member.
    /// @return True if member already submitted a public key, else false.
    function hasMemberSubmittedPublicKey(address _member)
        internal
        view
        returns (bool)
    {
        return submittedPublicKeys[_member].length != 0;
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
    // TODO: Add modifier to be able to run this function only when keep was
    // closed before.
    // TODO: Rename to `seizeMembersBonds` for consistency.
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
    /// @dev The function expects the signed digest to be calculated as a sha256 hash
    /// of the preimage: `sha256(_preimage))`.
    /// @param _v Signature's header byte: `27 + recoveryID`.
    /// @param _r R part of ECDSA signature.
    /// @param _s S part of ECDSA signature.
    /// @param _signedDigest Digest for the provided signature. Result of hashing
    /// the preimage with sha256.
    /// @param _preimage Preimage of the hashed message.
    /// @return True if fraud, error otherwise.
    function submitSignatureFraud(
        uint8 _v,
        bytes32 _r,
        bytes32 _s,
        bytes32 _signedDigest,
        bytes calldata _preimage
    ) external returns (bool _isFraud) {
        require(publicKey.length != 0, "Public key was not set yet");

        bytes32 calculatedDigest = sha256(_preimage);
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
    function sign(bytes32 _digest) external onlyOwner onlyWhenActive {
        require(publicKey.length != 0, "Public key was not set yet");
        require(!isSigningInProgress(), "Signer is busy");

        /* solium-disable-next-line */
        signingStartTimestamp = block.timestamp;

        digests[_digest] = true;
        digest = _digest;

        emit SignatureRequested(_digest);
    }

    /// @notice Checks if keep is currently awaiting a signature for the given digest.
    /// @dev Validates if the signing is currently in progress and compares provided
    /// digest with the one for which the latest signature was requested.
    /// @param _digest Digest for which to check if signature is being awaited.
    /// @return True if the digest is currently expected to be signed, else false.
    function isAwaitingSignature(bytes32 _digest) external view returns (bool) {
        return isSigningInProgress() && digest == _digest;
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

    /// @notice Closes keep when owner decides that they no longer need it.
    /// Releases bonds to the keep members. Keep can be closed only when
    /// there is no signing in progress or requested signing process has timed out.
    /// @dev The function can be called by the owner of the keep and only is the
    /// keep has not been closed already.
    function closeKeep() external onlyOwner onlyWhenActive {
        require(
            !isSigningInProgress() || hasSigningTimedOut(),
            "Requested signing has not timed out yet"
        );

        isActive = false;

        freeMembersBonds();

        emit KeepClosed();
    }

    /// @notice Returns bonds to the keep members.
    function freeMembersBonds() internal {
        for (uint256 i = 0; i < members.length; i++) {
            keepBonding.freeBond(members[i], uint256(address(this)));
        }
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
    /// ETH is sent to the beneficiary of each member. If the value cannot be
    /// divided evenly across the members submits the remainder to the last keep
    /// member.
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
            /* solium-disable-next-line security/no-call-value */
            tokenStaking.magpieOf(members[i]).call.value(dividend)("");
        }

        // Check if value has a remainder after dividing it across members.
        // If so send the remainder to the last member.
        uint256 remainder = msg.value.mod(memberCount);
        if (remainder > 0) {
            /* solium-disable-next-line security/no-call-value */
            tokenStaking.magpieOf(members[memberCount - 1]).call.value(
                remainder
            )("");
        }
    }

    /// @notice Distributes ERC20 token evenly across all keep members.
    /// The token is sent to the beneficiary of each member.
    /// @dev This works with any ERC20 token that implements a transferFrom
    /// function similar to the interface imported here from
    /// openZeppelin. This function only has authority over pre-approved
    /// token amount. We don't explicitly check for allowance, SafeMath
    /// subtraction overflow is enough protection. If the value cannot be
    /// divided evenly across the members submits the remainder to the last keep
    /// member.
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

        // Check if value has a remainder after dividing it across members.
        // If so send the remainder to the last member.
        uint256 remainder = _value.mod(memberCount);
        if (remainder > 0) {
            token.transferFrom(
                msg.sender,
                tokenStaking.magpieOf(members[memberCount - 1]),
                remainder
            );
        }

    }

    /// @notice Checks if the caller is a keep member.
    /// @dev Throws an error if called by any account other than one of the members.
    modifier onlyMember() {
        require(members.contains(msg.sender), "Caller is not the keep member");
        _;
    }

    /// @notice Checks if the keep is currently active.
    /// @dev Throws an error if called when the keep has been already closed.
    modifier onlyWhenActive() {
        require(isActive, "Keep is not active");
        _;
    }
}
