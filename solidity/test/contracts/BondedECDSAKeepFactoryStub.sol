pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeepFactory.sol";

/// @title Bonded ECDSA Keep Factory Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepFactoryStub is BondedECDSAKeepFactory {
    constructor(
        address sortitionPoolFactory,
        address tokenStaking,
        address keepBonding,
        address randomBeacon
    )
        public
        BondedECDSAKeepFactory(
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

    function registerMemberCandidateStub(address _application, address _operator) external {
        require(
            candidatesPools[_application] != address(0),
            "No pool found for the application"
        );

        BondedSortitionPool candidatesPool = BondedSortitionPool(
            candidatesPools[_application]
        );

        address operator = _operator;
        if (!candidatesPool.isOperatorInPool(operator)) {
            candidatesPool.joinPool(operator);
        }
    }
}
