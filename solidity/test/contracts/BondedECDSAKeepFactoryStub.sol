pragma solidity ^0.5.4;

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
    /// @return Created keep address.
    function stubOpenKeep(
        address _owner
    ) external payable returns (address keepAddress) {

        address[] memory members;
        keepAddress = createClone(masterBondedECDSAKeepAddress);
        BondedECDSAKeep keep = BondedECDSAKeep(keepAddress);
        keep.initialize(
            _owner,
            members,
            0,
            address(0),
            address(0)
        );
        keeps.push(keepAddress);
    }
}
