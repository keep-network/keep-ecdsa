pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./api/IECDSAKeepFactory.sol";
import "./utils/AddressArrayUtils.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@keep-network/sortition-pools/contracts/SortitionPool.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
contract ECDSAKeepFactory is IECDSAKeepFactory { // TODO: Rename to BondedECDSAKeepFactory
    using AddressArrayUtils for address payable[];
    using SafeMath for uint256;

    // List of keeps.
    ECDSAKeep[] keeps;

    //SortitionPool pool;

    // List of candidates to be selected as keep members. Once the candidate is
    // registered it remains on the list forever.
    // TODO: It's a temporary solution until we implement proper candidate
    // registration and member selection.
    address payable[] memberCandidates;

    // Notification that a new keep has been created.
    event ECDSAKeepCreated(
        address keepAddress,
        address payable[] members
    );

    /// @notice Register caller as a candidate to be selected as keep member.
    /// @dev If caller is already registered it returns without any changes.
    /// TODO: This is a simplified solution until we have proper registration
    /// and pool selection.
    function registerMemberCandidate() external {
        if (!memberCandidates.contains(msg.sender)) {
            memberCandidates.push(msg.sender);
        }
    }

    /// @notice Open a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress) {
        address payable[] memory _members = selectECDSAKeepMembers(_groupSize);

        ECDSAKeep keep = new ECDSAKeep(
            _owner,
            _members,
            _honestThreshold
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress, _members);
    }

    // TODO: Selection of ECDSA Keep members will be rewritten.

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations generates a group with only one member. Member
    /// is randomly selected from registered member candidates.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectECDSAKeepMembers(
        uint256 _groupSize
    ) internal view returns (address payable[] memory members){
        require(memberCandidates.length > 0, 'keep member candidates list is empty');

        _groupSize;

        members = new address payable[](1);

        // TODO: Use the random beacon for randomness.
        uint memberIndex = uint256(keccak256(abi.encodePacked(block.timestamp)))
            % memberCandidates.length;

        members[0] = memberCandidates[memberIndex];
    }
}
