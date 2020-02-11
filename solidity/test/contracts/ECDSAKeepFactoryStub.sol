pragma solidity ^0.5.4;

import "../../contracts/ECDSAKeepFactory.sol";

/// @title ECDSA Keep Factory Stub
/// @dev This contract is for testing purposes only.
contract ECDSAKeepFactoryStub is ECDSAKeepFactory {
    constructor(
        address sortitionPoolFactory,
        address tokenStaking,
        address keepBonding,
        address randomBeacon
    )
        public
        ECDSAKeepFactory(
            sortitionPoolFactory,
            tokenStaking,
            keepBonding,
            randomBeacon
        )
    {}

    // @dev Returns address of registered signer pool.
    function getSignerPool(address application) public view returns (address) {
        return candidatesPools[application];
    }

    function initialGroupSelectionSeed(uint256 _groupSelectionSeed) public {
        groupSelectionSeed = _groupSelectionSeed;
    }

    function getGroupSelectionSeed() public view returns (uint256) {
        return groupSelectionSeed;
    }
}
