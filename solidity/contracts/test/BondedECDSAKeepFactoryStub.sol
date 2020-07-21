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

    function addKeep(address keep) public {
        keeps.push(keep);
        /* solium-disable-next-line security/no-block-members*/
        keepOpenedTimestamp[keep] = block.timestamp;
    }
}
