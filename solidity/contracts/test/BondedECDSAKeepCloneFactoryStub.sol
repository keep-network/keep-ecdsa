pragma solidity 0.5.17;

import "../../contracts/BondedECDSAKeep.sol";
import "../../contracts/CloneFactory.sol";


/// @title Bonded ECDSA Keep Factory Stub using clone factory.
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepCloneFactory is CloneFactory {
    uint256 public minimumStake = 200000 * 1e18;

    address public masterBondedECDSAKeepAddress;
    bool public membersSlashed;

    constructor(address _masterBondedECDSAKeepAddress) public {
        masterBondedECDSAKeepAddress = _masterBondedECDSAKeepAddress;
    }

    event BondedECDSAKeepCreated(address keepAddress);

    function newKeep(
        address _owner,
        address[] calldata _members,
        uint256 _honestThreshold,
        uint256 _minimumStake,
        uint256 _stakeLockDuration,
        address _tokenStaking,
        address _keepBonding,
        address payable _keepFactory
    ) external payable returns (address keepAddress) {
        keepAddress = createClone(masterBondedECDSAKeepAddress);
        assert(isClone(masterBondedECDSAKeepAddress, keepAddress));

        BondedECDSAKeep keep = BondedECDSAKeep(keepAddress);
        keep.initialize(
            _owner,
            _members,
            _honestThreshold,
            _minimumStake,
            _stakeLockDuration,
            _tokenStaking,
            _keepBonding,
            _keepFactory
        );

        emit BondedECDSAKeepCreated(keepAddress);
    }
}
