pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "./api/IECDSAKeep.sol";
import "./utils/AddressArrayUtils.sol";

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeep is IECDSAKeep, Ownable {
    using AddressArrayUtils for address payable[];
    using SafeMath for uint256;

    // List of keep members' addresses.
    address payable[] internal members;
    // Minimum number of honest keep members required to produce a signature.
    uint256 honestThreshold;
    // Signer's ECDSA public key serialized to 64-bytes, where X and Y coordinates
    // are padded with zeros to 32-byte each.
    bytes publicKey;

    // Notification that a signer's public key was published for the keep.
    event PublicKeyPublished(
        bytes publicKey
    );

    // Notification that the keep was requested to sign a digest.
    event SignatureRequested(
        bytes32 digest
    );

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
        uint8   recoveryID
    );

    constructor(
        address _owner,
        address payable[] memory _members,
        uint256 _honestThreshold
    ) public {
        transferOwnership(_owner);
        members = _members;
        honestThreshold = _honestThreshold;
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

    /// @notice Calculates a signature over provided digest by the keep.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) external onlyOwner {
        emit SignatureRequested(_digest);
    }

    /// @notice Submits a signature calculated for the given digest.
    /// @param _digest Digest for which calculator was calculated.
    /// @param _r Calculated signature's R value.
    /// @param _s Calculated signature's S value.
    /// @param _recoveryID Calculated signature's recovery ID (one of {0, 1, 2, 3}).
    function submitSignature(
        bytes32 _digest,
        bytes32 _r,
        bytes32 _s,
        uint8 _recoveryID
    ) external onlyMember {
        require(_recoveryID < 4, "Recovery ID must be one of {0, 1, 2, 3}");

        // We add 27 to the recovery ID to align it with ethereum and bitcoin
        // protocols where 27 is added to recovery ID to indicate usage of
        // uncompressed public keys.
        uint8 _v = 27 + _recoveryID;

        // Validate signature.
        require(
            publicKeyToAddress(publicKey) == ecrecover(_digest, _v, _r, _s),
            "Invalid signature"
        );

        emit SignatureSubmitted(_digest, _r, _s, _recoveryID);
    }

    /// @notice Checks if the caller is a keep member.
    /// @dev Throws an error if called by any account other than one of the members.
    modifier onlyMember() {
        require(members.contains(msg.sender), "Caller is not the keep member");
        _;
    }

    /// @notice Coverts a public key to an ethereum address.
    /// @param _publicKey Public key provided as 64-bytes concatenation of
    /// X and Y coordinates (32-bytes each).
    /// @return Ethereum address.
    function publicKeyToAddress(bytes memory _publicKey) internal pure returns (address) {
        // We hash the public key and then truncate last 20 bytes of the digest
        // which is the ethereum address.
        return address(uint160(uint256(keccak256(_publicKey))));
    }

    /// @notice Distributes eth evnly across all keep members
    /// @dev    Only the value passed to this function will be distributed
    /// @return true if the function completes execution
    function distributeETHToMembers() public payable {
        uint256 memberCount = members.length;
        uint256 dividend = msg.value.div(memberCount);

        require(dividend > 0, "must send value with TX");

        for(uint8 i = 0; i < memberCount; i++){
            members[i].transfer(dividend);
        }
    }

    /// @notice Distributes ERC20 token evnly across all keep members
    /// @dev    This works with any ERC20 token that implements a transferFrom
    ///         function similar to the interface imported here from
    ///         openZeppelin. This function only has authority over pre-approved
    ///         token amount. We don't explicitly check for allowance, SafeMath
    ///         subtraction overflow is enough protection.
    /// @param _tokenAddress Address of the ERC20 token to distribute
    /// @param _value        Amount of ERC20 token to destribute
    /// @return              True if the function completes execution
    function distributeERC20ToMembers(address _tokenAddress, uint256 _value) public {
        IERC20 token = IERC20(_tokenAddress);

        uint256 memberCount = members.length;
        uint256 dividend = _value.div(memberCount);

        require(dividend > 0, "value must be non-zero");

        for(uint8 i = 0; i < memberCount; i++){
            token.transferFrom(msg.sender, members[i], dividend);
        }
    }
}
