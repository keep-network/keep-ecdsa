pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./utils/AddressArrayUtils.sol";

/// @title ECDSA Keep
/// @notice Contract reflecting an ECDSA keep.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeep is Ownable {
    using AddressArrayUtils for address[];

    // List of keep members' addresses.
    address[] internal members;
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
    // was used for signature calculation and a signature in a form of R, S, V
    // values.
    event SignatureSubmitted(
        bytes32 digest,
        bytes32 r,
        bytes32 s,
        uint8   v
    );

    constructor(
        address _owner,
        address[] memory _members,
        uint256 _honestThreshold
    ) public {
        transferOwnership(_owner);
        members = _members;
        honestThreshold = _honestThreshold;
    }

    /// @notice Set a signer's public key for the keep.
    /// @dev Stub implementations.
    /// @param _publicKey Signer's public key.
    function setPublicKey(bytes memory _publicKey) public onlyMember {
        require(_publicKey.length == 64, "Public key must be 64 bytes long");
        publicKey = _publicKey;
        emit PublicKeyPublished(_publicKey);
    }

    /// @notice Returns the keep signer's public key.
    /// @return Signer's public key.
    function getPublicKey() public view returns (bytes memory) {
       return publicKey;
    }

    /// @notice Calculates a signature over provided digest by the keep.
    /// @param _digest Digest to be signed.
    function sign(bytes32 _digest) public onlyOwner {
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
    ) public onlyMember {
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

        emit SignatureSubmitted(_digest, _r, _s, _v);
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
}
