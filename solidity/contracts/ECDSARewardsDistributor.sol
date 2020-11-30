/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓▌        ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "@keep-network/keep-core/contracts/utils/BytesLib.sol";
import "openzeppelin-solidity/contracts/token/ERC20/SafeERC20.sol";
import "@keep-network/keep-core/contracts/KeepToken.sol";
import "@openzeppelin/contracts/cryptography/MerkleProof.sol";

/// @title ECDSA Rewards distributor
/// @notice This contract is based on the Uniswap's Merkle Distributor
/// https://github.com/Uniswap/merkle-distributor with some modifications:
/// - added a map of merkle root keys with the amount of KEEP (value) that will
///   be allocated for those merkle roots
/// - added receiveApproval() function that will be called each time to allocate
///   new KEEP rewards for a given merkle root. Merkle root is going to be generated
///   regulary (ex. every week) and it is also means that an interval for that
///   merkle root has passed
/// - changed code accordingly to process claimed rewards using a map of merkle
///   roots
contract ECDSARewardsDistributor is Ownable {
    using SafeERC20 for KeepToken;
    using BytesLib for bytes;

    KeepToken public token;

    // This event is triggered whenever a call to #claim succeeds.
    event RewardsClaimed(uint256 index, address account, uint256 amount);
    // This event is triggered whenever rewards are allocated.
    event RewardsAllocated(bytes32 merkleRoot, uint256 amount);

    // Merkle root -> total amount for distribution for a given interval.
    mapping(bytes32 => uint256) private merkleRoots;
    // Bytes32 key is a merkle root and the value is a packed array of booleans.
    mapping(bytes32 => mapping(uint256 => uint256)) private claimedBitMap;

    constructor(address _token) public {
        token = KeepToken(_token);
    }

    /// Call receiveApproval to allocate amount of KEEP for a given merkle root.
    /// @param _from The original sender of the tokens.
    /// @param _amount The amount of KEEP tokens to fund.
    /// @param _token The KEEP token to fund the rewards in.
    /// @param _extraData Merkle root (32 bytes).
    function receiveApproval(
        address _from,
        uint256 _amount,
        address _token,
        bytes memory _extraData
    ) public {
        require(IERC20(_token) == token, "Unsupported token");
        require(_extraData.length == 32, "Wrong lenght of merkle root");

        token.safeTransferFrom(_from, address(this), _amount);

        bytes32 merkleRoot = _extraData.toBytes32();

        merkleRoots[merkleRoot] = _amount;

        emit RewardsAllocated(merkleRoot, _amount);
    }

    function isClaimed(uint256 index, bytes32 merkleRoot) public view returns (bool) {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        uint256 claimedWord = claimedBitMap[merkleRoot][claimedWordIndex];
        uint256 mask = (1 << claimedBitIndex);
        return claimedWord & mask == mask;
    }

    function _setClaimed(uint256 index, bytes32 merkleRoot) private {
        uint256 claimedWordIndex = index / 256;
        uint256 claimedBitIndex = index % 256;
        claimedBitMap[merkleRoot][claimedWordIndex] = claimedBitMap[merkleRoot][claimedWordIndex] | (1 << claimedBitIndex);
    }

    function claim(uint256 index, address account, uint256 amount, bytes32 merkleRoot, bytes32[] calldata merkleProof) external {
        require(!isClaimed(index, merkleRoot), 'Rewards already claimed');

        // Verify the merkle proof.
        bytes32 node = keccak256(abi.encodePacked(index, account, amount));

        require(MerkleProof.verify(merkleProof, merkleRoot, node), 'Invalid proof');

        // Mark it claimed and send the token.
        _setClaimed(index, merkleRoot);
        require(IERC20(token).transfer(account, amount), 'Transfer failed');

        // Update KEEP amount for the given merkleRoot
        merkleRoots[merkleRoot] -= amount;

        emit RewardsClaimed(index, account, amount);
    }
}