pragma solidity 0.5.17;

import "../../contracts/BondedECDSAKeep.sol";

contract RewardsKeepStub is BondedECDSAKeep {
    function setTimestamp(uint256 time) external {
        keyGenerationStartTimestamp = time;
    }

    function close() external {
        markAsClosed();
    }

    function terminate() external {
        markAsTerminated();
    }
}
