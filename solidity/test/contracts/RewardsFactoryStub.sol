pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeepFactory.sol";

/// @title Bonded ECDSA Keep Factory Stub
/// @dev This contract is for testing purposes only.
contract RewardsFactoryStub is BondedECDSAKeepFactory {
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

    /// @notice Opens new keeps with the provided arbitrary timestamps.
    /// @param timestamps Arbitrary timestamps, in ascending order.
    function openSyntheticKeeps(
        address[] calldata members,
        uint256[] calldata timestamps
    ) external {
        for (uint256 i = 0; i < timestamps.length; i++) {
            require(
                i == 0 || timestamps[i] >= timestamps[i-1],
                "provided timestamps not in order"
            );
            address keepAddress = createClone(masterBondedECDSAKeepAddress);
            BondedECDSAKeep keep = BondedECDSAKeep(keepAddress);
            keep.initialize(
                msg.sender,
                members,
                members.length,
                address(tokenStaking),
                address(keepBonding)
            );
            keeps.push(keepAddress);
            creationTime[keepAddress] = timestamps[i];
        }
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
