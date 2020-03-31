pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeep.sol";

/// @title Bonded ECDSA Keep Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepStub is BondedECDSAKeep {
    function publicSlashSignerStakes() public {
        slashSignerStakes();
    }

    function publicMarkAsClosed() public {
        markAsClosed();
    }

    function publicMarkAsTerminated() public {
        markAsTerminated();
    }
}
