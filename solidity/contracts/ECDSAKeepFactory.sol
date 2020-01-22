pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./utils/AddressArrayUtils.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating ECDSA keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeepFactory {
    using AddressArrayUtils for address payable[];

    // List of keeps.
    ECDSAKeep[] keeps;

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
    /// and group selection.
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

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations generates a group with only one member. Member
    /// is randomly selected from registered member candidates.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectECDSAKeepMembers(
        uint256 _groupSize
    ) internal view returns (address payable[] memory members){
         members = new address payable[](_groupSize);

        // TODO: Use the random beacon for randomness.
        uint firstIndex = uint256(keccak256(abi.encodePacked(block.timestamp)))
            % memberCandidates.length;

        // TODO: Temporary solution until group selection protocol is implemented.
        uint nextIndex = firstIndex;
        for (uint i = 0; i < _groupSize; i++) {
            members[i] = memberCandidates[nextIndex];
            nextIndex++;
            nextIndex = nextIndex % memberCandidates.length;
        }
    }
}
