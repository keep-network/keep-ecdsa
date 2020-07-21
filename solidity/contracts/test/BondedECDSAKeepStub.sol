pragma solidity 0.5.17;

import "../../contracts/BondedECDSAKeep.sol";


/// @title Bonded ECDSA Keep Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepStub is BondedECDSAKeep {
    function publicMarkAsClosed() public {
        markAsClosed();
    }

    function publicMarkAsTerminated() public {
        markAsTerminated();
    }

    function isFradulentPreimageSet(bytes memory preimage) public view returns (bool) {
        return fraudulentPreimages[preimage];
    }

    function setMemberStake(uint256 stake) public {
        memberStake = stake;
    }
}
