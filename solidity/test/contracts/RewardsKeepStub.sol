pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeep.sol";

contract RewardsKeepStub is BondedECDSAKeep {
    function setTimestamp(uint256 time) external {
        keyGenerationStartTimestamp = time;
    }
}
