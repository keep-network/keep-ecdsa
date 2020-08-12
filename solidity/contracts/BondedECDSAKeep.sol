/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓         ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "./AbstractECDSAKeep.sol";
import "./api/IBondedECDSAKeep.sol";

import "./BondedECDSAKeepFactory.sol";
import "./KeepBonding.sol";

import "@keep-network/keep-core/contracts/TokenStaking.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Bonded ECDSA Keep
/// @notice ECDSA keep with additional signer bond requirement.
/// @dev This contract is used as a master contract for clone factory in
/// BondedECDSAKeepFactory as per EIP-1167. It should never be removed after
/// initial deployment as this will break functionality for all created clones.
contract BondedECDSAKeep is IBondedECDSAKeep, AbstractECDSAKeep {
    using AddressArrayUtils for address[];
    using SafeMath for uint256;

    // Stake that was required from each keep member on keep creation.
    // The value is used for keep members slashing.
    uint256 public memberStake;

    // Emitted when KEEP token slashing failed when submitting signature
    // fraud proof. In practice, this situation should never happen but we want
    // to be very explicit in this contract and protect the owner that even if
    // it happens, the transaction submitting fraud proof is not going to fail
    // and keep owner can seize and liquidate bonds in the same transaction.
    event SlashingFailed();

    TokenStaking tokenStaking;
    KeepBonding keepBonding;
    BondedECDSAKeepFactory keepFactory;

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

    /// @notice Closes keep when owner decides that they no longer need it.
    /// Releases bonds to the keep members. Keep can be closed only when
    /// there is no signing in progress or requested signing process has timed out.
    /// @dev The function can be called only by the owner of the keep and only
    /// if the keep has not been already closed.
    function closeKeep() external onlyOwner onlyWhenActive {
        markAsClosed();
        unlockMemberStakes();
        freeMembersBonds();
    }

    /// @notice Seizes the signers' ETH bonds. After seizing bonds keep is
    /// closed so it will no longer respond to signing requests. Bonds can be
    /// seized only when there is no signing in progress or requested signing
    /// process has timed out. This function seizes all of signers' bonds.
    /// The application may decide to return part of bonds later after they are
    /// processed using returnPartialSignerBonds function.
    function seizeSignerBonds() external onlyOwner onlyWhenActive {
        terminateKeep();

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
    /// not first approved via a call to sign. If fraud is detected it tries to
    /// slash members' KEEP tokens. For each keep member tries slashing amount
    /// equal to the member stake set by the factory when keep was created.
    /// @dev The function expects the signed digest to be calculated as a sha256
    /// hash of the preimage: `sha256(_preimage))`. The function reverts if the
    /// signature is not fraudulent. The function does not revert if KEEP slashing
    /// failed but emits an event instead. In practice, KEEP slashing should
    /// never fail.
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
    ) external onlyWhenActive returns (bool _isFraud) {
        bool isFraud = checkSignatureFraud(
            _v,
            _r,
            _s,
            _signedDigest,
            _preimage
        );

        require(isFraud, "Signature is not fraudulent");

        if (!fraudulentPreimages[_preimage]) {
            /* solium-disable-next-line */
            (bool success, ) = address(tokenStaking).call(
                abi.encodeWithSignature(
                    "slash(uint256,address[])",
                    memberStake,
                    members
                )
            );

            fraudulentPreimages[_preimage] = true;

            // Should never happen but we want to protect the owner and make sure the
            // fraud submission transaction does not fail so that the owner can
            // seize and liquidate bonds in the same transaction.
            if (!success) {
                emit SlashingFailed();
            }
        }

        return isFraud;
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
        initializeECDSAKeep(_owner, _members, _honestThreshold);

        memberStake = _memberStake;
        tokenStaking = TokenStaking(_tokenStaking);
        keepBonding = KeepBonding(_keepBonding);
        keepFactory = BondedECDSAKeepFactory(_keepFactory);

        tokenStaking.claimDelegatedAuthority(_keepFactory);

        for (uint256 i = 0; i < _members.length; i++) {
            tokenStaking.lockStake(_members[i], _stakeLockDuration);
        }
    }

    /// @notice Terminates the keep.
    /// Keep can be marked as terminated only when there is no signing in progress
    /// or the requested signing process has timed out.
    function terminateKeep() internal {
        unlockMemberStakes();
        markAsTerminated();
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

    /// @notice Gets the beneficiary for the specified member address.
    /// @param _member Member address.
    /// @return Beneficiary address.
    function beneficiaryOf(address _member)
        internal
        view
        returns (address payable)
    {
        return tokenStaking.beneficiaryOf(_member);
    }
}
