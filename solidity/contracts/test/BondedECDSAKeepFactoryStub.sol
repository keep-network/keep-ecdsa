pragma solidity 0.5.17;

import "../../contracts/BondedECDSAKeepFactory.sol";

/// @title Bonded ECDSA Keep Factory Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepFactoryStub is BondedECDSAKeepFactory {
    constructor(
        address masterBondedECDSAKeepAddress,
        address sortitionPoolFactory,
        address tokenStaking,
        address keepBonding,
        address randomBeacon
    )
        public
        BondedECDSAKeepFactory(
            masterBondedECDSAKeepAddress,
            sortitionPoolFactory,
            tokenStaking,
            keepBonding,
            randomBeacon
        )
    {}

    function initialGroupSelectionSeed(uint256 _groupSelectionSeed) public {
        groupSelectionSeed = _groupSelectionSeed;
    }

    function getGroupSelectionSeed() public view returns (uint256) {
        return groupSelectionSeed;
    }

    /// @notice Opens a new ECDSA keep.
    /// @param _owner Address of the keep owner.
    /// @param _members Keep members.
    /// @param _creationTimestamp Keep creation timestamp.
    ///
    /// @return Created keep address.
    function stubOpenKeep(
        address _owner,
        address[] memory _members,
        uint256 _creationTimestamp
    ) public returns (address keepAddress) {
        keepAddress = createClone(masterKeepAddress);
        BondedECDSAKeep keep = BondedECDSAKeep(keepAddress);
        keep.initialize(
            _owner,
            _members,
            0,
            0,
            0,
            address(tokenStaking),
            address(keepBonding),
            address(this)
        );
        keeps.push(keepAddress);
        keepOpenedTimestamp[keepAddress] = _creationTimestamp;
    }

    /// @notice Mocks opening `_numberOfKeeps` with the opening timestamp of
    /// the first keep starting at `_firstKeepCreationTimestamp`.
    /// IMPORTANT! This function does not actually create new keep instances!
    /// It would not be possible to create 100 keeps in one block because of
    /// gas limits.
    function stubBatchOpenFakeKeeps(
        uint256 _numberOfKeeps,
        uint256 _firstKeepCreationTimestamp
    ) public {
        for (uint256 i = 0; i < _numberOfKeeps; i++) {
            address keepAddress = address(block.timestamp.add(i));
            keeps.push(keepAddress);
            keepOpenedTimestamp[keepAddress] = _firstKeepCreationTimestamp.add(
                i
            );
        }
    }
}
