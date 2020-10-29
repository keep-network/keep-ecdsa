pragma solidity 0.5.17;

import "../../contracts/fully-backed/FullyBackedECDSAKeep.sol";
import "../../contracts/CloneFactory.sol";

import {
    AuthorityDelegator
} from "@keep-network/keep-core/contracts/Authorizations.sol";

/// @title Fully Backed Bonded ECDSA Keep Factory Stub using clone factory.
/// @dev This contract is for testing purposes only.
contract FullyBackedECDSAKeepCloneFactoryStub is
    CloneFactory,
    AuthorityDelegator
{
    address public masterKeepAddress;

    mapping(address => uint256) public banKeepMembersCalledCount;

    constructor(address _masterKeepAddress) public {
        masterKeepAddress = _masterKeepAddress;
    }

    event FullyBackedECDSAKeepCreated(address keepAddress);

    function newKeep(
        address _owner,
        address[] calldata _members,
        uint256 _honestThreshold,
        address _keepBonding,
        address payable _keepFactory
    ) external payable returns (address keepAddress) {
        keepAddress = createClone(masterKeepAddress);
        assert(isClone(masterKeepAddress, keepAddress));

        FullyBackedECDSAKeep keep = FullyBackedECDSAKeep(keepAddress);
        keep.initialize(
            _owner,
            _members,
            _honestThreshold,
            _keepBonding,
            _keepFactory
        );

        emit FullyBackedECDSAKeepCreated(keepAddress);
    }

    function __isRecognized(address _delegatedAuthorityRecipient)
        external
        returns (bool)
    {
        return true;
    }

    function banKeepMembers() public {
        banKeepMembersCalledCount[msg.sender]++;
    }
}
