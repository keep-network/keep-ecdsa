pragma solidity 0.5.17;

import "./KeepBonding.sol";
import "./api/IBondedECDSAKeep.sol";
import "./BondedECDSAKeepFactory.sol";

import "@keep-network/keep-core/contracts/TokenStaking.sol";
import "@keep-network/keep-core/contracts/utils/AddressArrayUtils.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "openzeppelin-solidity/contracts/token/ERC20/SafeERC20.sol";


/// @title Bonded ECDSA Keep
/// @notice ECDSA keep with additional signer bond requirement.
/// @dev This contract is used as a master contract for clone factory in
/// BondedECDSAKeepFactory as per EIP-1167. It should never be removed after
/// initial deployment as this will break functionality for all created clones.
contract BondedECDSAKeep is IBondedECDSAKeep {
    using AddressArrayUtils for address[];
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    // Status of the keep.
    // Active means the keep is active.
    // Closed means the keep was closed happily.
    // Terminated means the keep was closed due to misbehavior.
    enum Status {Active, Closed, Terminated}

    // Flags execution of contract initialization.
    bool isInitialized;

    // Address of the keep's owner.
    address private owner;

    // List of keep members' addresses.
    address[] internal members;

    // Minimum number of honest keep members required to produce a signature.
    uint256 honestThreshold;

    // Stake that was required from each keep member on keep creation.
    // The value is used for keep members slashing.
    uint256 public memberStake;

    // Keep's ECDSA public key serialized to 64-bytes, where X and Y coordinates
    // are padded with zeros to 32-byte each.
    bytes publicKey;

    // Latest digest requested to be signed. Used to validate submitted signature.
    bytes32 public digest;

    // Map of all digests requested to be signed. Used to validate submitted signature.
    mapping(bytes32 => bool) digests;

    // Timeout for the keep public key to appear on the chain. Time is counted
    // from the moment keep has been created.
    uint256 public constant keyGenerationTimeout = 150 * 60; // 2.5h in seconds

    // The timestamp at which keep has been created and key generation process
    // started.
    uint256 internal keyGenerationStartTimestamp;

    // Timeout for a signature to appear on the chain. Time is counted from the
    // moment signing request occurred.
    uint256 public constant signingTimeout = 90 * 60; // 1.5h in seconds

    // The timestamp at which signing process started. Used also to track if
    // signing is in progress. When set to `0` indicates there is no
    // signing process in progress.
    uint256 internal signingStartTimestamp;

    // Map stores public key by member addresses. All members should submit the
    // same public key.
    mapping(address => bytes) submittedPublicKeys;

    // Map stores amount of wei stored in the contract for each member address.
    mapping(address => uint256) memberETHBalances;

    // The current status of the keep.
    // If the keep is Active members monitor it and support requests from the
    // keep owner.
    // If the owner decides to close the keep the flag is set to Closed.
    // If the owner seizes member bonds the flag is set to Terminated.
    Status internal status;

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

    // Notification that keep's ECDSA public key has been successfully established.
    event PublicKeyPublished(bytes publicKey);

    // Notification that ETH reward has been distributed to keep members.
    event ETHRewardDistributed();

    // Notification that ERC20 reward has been distributed to keep members.
    event ERC20RewardDistributed();

    // Notification that the keep was closed by the owner.
    // Members no longer need to support this keep.
    event KeepClosed();

    // Notification that the keep has been terminated by the owner.
    // Members no longer need to support this keep.
    event KeepTerminated();

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
    BondedECDSAKeepFactory keepFactory;

    /// @notice Returns keep's ECDSA public key.
    /// @return Keep's ECDSA public key.
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

    /// @notice Submits a public key to the keep.
    /// @dev Public key is published successfully if all members submit the same
    /// value. In case of conflicts with others members submissions it will emit
    /// `ConflictingPublicKeySubmitted` event. When all submitted keys match
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

    /// @notice Closes keep when owner decides that they no longer need it.
    /// Releases bonds to the keep members. Keep can be closed only when
    /// there is no signing in progress or requested signing process has timed out.
    /// @dev The function can be called only by the owner of the keep and only
    /// if the keep has not been already closed.
    function closeKeep() external onlyOwner onlyWhenActive {
        markAsClosed();
        freeMembersBonds();
    }

    /// @notice Seizes the signers' ETH bonds. After seizing bonds keep is
    /// closed so it will no longer respond to signing requests. Bonds can be
    /// seized only when there is no signing in progress or requested signing
    /// process has timed out. This function seizes all of signers' bonds.
    /// The application may decide to return part of bonds later after they are
    /// processed using returnPartialSignerBonds function.
    function seizeSignerBonds() external onlyOwner onlyWhenActive {
        markAsTerminated();

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
                address(uint160(owner))
            );
        }
    }

    /// @notice Returns partial signer's ETH bonds to the pool as an unbounded
    /// value. This function is called after bonds have been seized and processed
    /// by the privileged application after calling seizeSignerBonds function.
    /// It is entirely up to the application if a part of signers' bonds is
    /// returned. The application may decide for that but may also decide to
    /// seize bonds and do not return anything.
    function returnPartialSignerBonds() external payable {
        uint256 memberCount = members.length;
        uint256 bondPerMember = msg.value.div(memberCount);

        require(bondPerMember > 0, "Partial signer bond must be non-zero");

        for (uint16 i = 0; i < memberCount - 1; i++) {
            keepBonding.deposit.value(bondPerMember)(members[i]);
        }

        // Transfer of dividend for the last member. Remainder might be equal to
        // zero in case of even distribution or some small number.
        uint256 remainder = msg.value.mod(memberCount);
        keepBonding.deposit.value(bondPerMember.add(remainder))(
            members[memberCount - 1]
        );
    }

    /// @notice Submits a fraud proof for a valid signature from this keep that was
    /// not first approved via a call to sign. If fraud is detected it slashes
    /// members' KEEP tokens.
    /// @dev The function expects the signed digest to be calculated as a sha256
    /// hash of the preimage: `sha256(_preimage))`. The function reverts if the
    /// signature is not fraudulent.
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
        bool isFraud = checkSignatureFraud(
            _v,
            _r,
            _s,
            _signedDigest,
            _preimage
        );

        require(isFraud, "Signature is not fraudulent");

        slashSignerStakes();

        return isFraud;
    }

    /// @notice Distributes ETH reward evenly across all keep signer beneficiaries.
    /// If the value cannot be divided evenly across all signers, it sends the
    /// remainder to the last keep signer.
    /// @dev Only the value passed to this function is distributed. This
    /// function does not transfer the value to beneficiaries accounts; instead
    /// it holds the value in the contract until withdraw function is called for
    /// the specific signer.
    function distributeETHReward() external payable {
        uint256 memberCount = members.length;
        uint256 dividend = msg.value.div(memberCount);

        require(dividend > 0, "Dividend value must be non-zero");

        for (uint16 i = 0; i < memberCount - 1; i++) {
            memberETHBalances[members[i]] += dividend;
        }

        // Give the dividend to the last signer. Remainder might be equal to
        // zero in case of even distribution or some small number.
        uint256 remainder = msg.value.mod(memberCount);
        memberETHBalances[members[memberCount - 1]] += dividend.add(remainder);

        emit ETHRewardDistributed();
    }

    /// @notice Distributes ERC20 reward evenly across all keep signer beneficiaries.
    /// @dev This works with any ERC20 token that implements a transferFrom
    /// function similar to the interface imported here from
    /// OpenZeppelin. This function only has authority over pre-approved
    /// token amount. We don't explicitly check for allowance, SafeMath
    /// subtraction overflow is enough protection. If the value cannot be
    /// divided evenly across the signers, it submits the remainder to the last
    /// keep signer.
    /// @param _tokenAddress Address of the ERC20 token to distribute.
    /// @param _value Amount of ERC20 token to distribute.
    function distributeERC20Reward(address _tokenAddress, uint256 _value)
        external
    {
        IERC20 token = IERC20(_tokenAddress);

        uint256 memberCount = members.length;
        uint256 dividend = _value.div(memberCount);

        require(dividend > 0, "Dividend value must be non-zero");

        for (uint16 i = 0; i < memberCount - 1; i++) {
            token.safeTransferFrom(
                msg.sender,
                tokenStaking.magpieOf(members[i]),
                dividend
            );
        }

        // Transfer of dividend for the last member. Remainder might be equal to
        // zero in case of even distribution or some small number.
        uint256 remainder = _value.mod(memberCount);
        token.safeTransferFrom(
            msg.sender,
            tokenStaking.magpieOf(members[memberCount - 1]),
            dividend.add(remainder)
        );

        emit ERC20RewardDistributed();
    }

    /// @notice Gets current amount of ETH hold in the keep for the member.
    /// @param _member Keep member address.
    /// @return Current balance in wei.
    function getMemberETHBalance(address _member)
        external
        view
        returns (uint256)
    {
        return memberETHBalances[_member];
    }

    /// @notice Withdraws amount of ether hold in the keep for the member.
    /// The value is sent to the beneficiary of the specific member.
    /// @param _member Keep member address.
    function withdraw(address _member) external {
        uint256 value = memberETHBalances[_member];

        require(value > 0, "No funds to withdraw");

        memberETHBalances[_member] = 0;

        /* solium-disable-next-line security/no-call-value */
        (bool success, ) = tokenStaking.magpieOf(_member).call.value(value)("");

        require(success, "Transfer failed");
    }

    /// @notice Gets the owner of the keep.
    /// @return Address of the keep owner.
    function getOwner() external view returns (address) {
        return owner;
    }

    /// @notice Gets the timestamp the keep was opened at.
    /// @return Timestamp the keep was opened at.
    function getOpenedTimestamp() external view returns (uint256) {
        return keyGenerationStartTimestamp;
    }

    /// @notice Initialization function.
    /// @dev We use clone factory to create new keep. That is why this contract
    /// doesn't have a constructor. We provide keep parameters for each instance
    /// function after cloning instances from the master contract.
    /// @param _owner Address of the keep owner.
    /// @param _members Addresses of the keep members.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _memberStake Stake required from each keep member.
    /// @param _stakeLockDuration Stake lock duration in seconds.
    /// @param _tokenStaking Address of the TokenStaking contract.
    /// @param _keepBonding Address of the KeepBonding contract.
    /// @param _keepFactory Address of the BondedECDSAKeepFactory that created
    /// this keep.
    function initialize(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold,
        uint256 _memberStake,
        uint256 _stakeLockDuration,
        address _tokenStaking,
        address _keepBonding,
        address payable _keepFactory
    ) public {
        require(!isInitialized, "Contract already initialized");

        owner = _owner;
        members = _members;
        honestThreshold = _honestThreshold;
        memberStake = _memberStake;
        tokenStaking = TokenStaking(_tokenStaking);
        keepBonding = KeepBonding(_keepBonding);
        keepFactory = BondedECDSAKeepFactory(_keepFactory);
        status = Status.Active;
        isInitialized = true;

        tokenStaking.claimDelegatedAuthority(_keepFactory);

        for (uint256 i = 0; i < _members.length; i++) {
            tokenStaking.lockStake(_members[i], _stakeLockDuration);
        }

        /* solium-disable-next-line security/no-block-members*/
        keyGenerationStartTimestamp = block.timestamp;
    }

    /// @notice Returns true if the keep is active.
    /// @return true if the keep is active, false otherwise.
    function isActive() public view returns (bool) {
        return status == Status.Active;
    }

    /// @notice Returns true if the keep is closed and members no longer support
    /// this keep.
    /// @return true if the keep is closed, false otherwise.
    function isClosed() public view returns (bool) {
        return status == Status.Closed;
    }

    /// @notice Returns true if the keep has been terminated.
    /// Keep is terminated when bonds are seized and members no longer support
    /// this keep.
    /// @return true if the keep has been terminated, false otherwise.
    function isTerminated() public view returns (bool) {
        return status == Status.Terminated;
    }

    /// @notice Returns members of the keep.
    /// @return List of the keep members' addresses.
    function getMembers() public view returns (address[] memory) {
        return members;
    }

    /// @notice Checks a fraud proof for a valid signature from this keep that was
    /// not first approved via a call to sign.
    /// @dev The function expects the signed digest to be calculated as a sha256 hash
    /// of the preimage: `sha256(_preimage))`. The digest is verified against the
    /// preimage to ensure the security of the ECDSA protocol. Verifying just the
    /// signature and the digest is not enough and leaves the possibility of the
    /// the existential forgery. If digest and preimage verification fails the
    /// function reverts.
    /// Reverts if a public key has not been set for the keep yet.
    /// @param _v Signature's header byte: `27 + recoveryID`.
    /// @param _r R part of ECDSA signature.
    /// @param _s S part of ECDSA signature.
    /// @param _signedDigest Digest for the provided signature. Result of hashing
    /// the preimage with sha256.
    /// @param _preimage Preimage of the hashed message.
    /// @return True if fraud, false otherwise.
    function checkSignatureFraud(
        uint8 _v,
        bytes32 _r,
        bytes32 _s,
        bytes32 _signedDigest,
        bytes memory _preimage
    ) public view returns (bool _isFraud) {
        require(publicKey.length != 0, "Public key was not set yet");

        bytes32 calculatedDigest = sha256(_preimage);
        require(
            _signedDigest == calculatedDigest,
            "Signed digest does not match double sha256 hash of the preimage"
        );

        bool isSignatureValid = publicKeyToAddress(publicKey) ==
            ecrecover(_signedDigest, _v, _r, _s);

        // Check if the signature is valid but was not requested.
        return isSignatureValid && !digests[_signedDigest];
    }

    /// @notice Returns true if the ongoing key generation process timed out.
    /// @dev There is a certain timeout for keep public key to be produced and
    /// appear on the chain, see `keyGenerationTimeout`.
    function hasKeyGenerationTimedOut() public view returns (bool) {
        /* solium-disable-next-line */
        return
            block.timestamp >
            keyGenerationStartTimestamp + keyGenerationTimeout;
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

    /// @notice Slashes keep members' KEEP tokens. For each keep member it slashes
    /// amount equal to the member stake set by the factory when keep was created.
    function slashSignerStakes() internal onlyWhenActive {
        tokenStaking.slash(memberStake, members);
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

    /// @notice Marks the keep as closed.
    /// Keep can be marked as closed only when there is no signing in progress
    /// or the requested signing process has timed out.
    function markAsClosed() internal {
        require(
            !isSigningInProgress() || hasSigningTimedOut(),
            "Requested signing has not timed out yet"
        );

        unlockMemberStakes();

        status = Status.Closed;
        emit KeepClosed();
    }

    /// @notice Marks the keep as terminated.
    /// Keep can be marked as terminated only when there is no signing in progress
    /// or the requested signing process has timed out.
    function markAsTerminated() internal {
        require(
            !isSigningInProgress() || hasSigningTimedOut(),
            "Requested signing has not timed out yet"
        );

        unlockMemberStakes();

        status = Status.Terminated;
        emit KeepTerminated();
    }

    /// @notice Releases locks the keep had previously placed on the members'
    /// token stakes.
    function unlockMemberStakes() internal {
        for (uint256 i = 0; i < members.length; i++) {
            tokenStaking.unlockStake(members[i]);
        }
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

    /// @notice Checks if the caller is the keep's owner.
    /// @dev Throws an error if called by any account other than owner.
    modifier onlyOwner() {
        require(owner == msg.sender, "Caller is not the keep owner");
        _;
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
        require(isActive(), "Keep is not active");
        _;
    }
}
