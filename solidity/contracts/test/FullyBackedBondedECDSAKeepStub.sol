pragma solidity 0.5.17;

import "../../contracts/fully-backed/FullyBackedBondedECDSAKeep.sol";

/// @title Fully Backed Bonded ECDSA Keep Stub
/// @dev This contract is for testing purposes only.
contract FullyBackedBondedECDSAKeepStub is FullyBackedBondedECDSAKeep {
    function publicMarkAsClosed() public {
        markAsClosed();
    }

    function publicMarkAsTerminated() public {
        markAsTerminated();
    }

    function publicSlashForSignatureFraud() public {
        slashForSignatureFraud();
    }

    function isFradulentPreimageSet(bytes memory preimage)
        public
        view
        returns (bool)
    {
        return fraudulentPreimages[preimage];
    }
}
