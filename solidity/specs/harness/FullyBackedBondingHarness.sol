pragma solidity 0.5.17;

import "../../contracts/fully-backed/FullyBackedBonding.sol";

contract FullyBackedBondingHarness is FullyBackedBonding {
    // total value in wei deposited by operators.
    mapping(address => uint256) public everDeposited;
    // mapping from operator,holder,referenceId to BondID
    mapping(address => mapping(address => mapping(uint256 => bytes32)))
        public operatorHolderRefToBondID;

    address payable public otherBeneficiary;

    constructor(KeepRegistry _keepRegistry, uint256 _initializationPeriod)
        public
        FullyBackedBonding(_keepRegistry, _initializationPeriod)
    {}

    function getBondID(
        address operator,
        address holder,
        uint256 referenceID
    ) internal view returns (bytes32) {

            bytes32 bondID
         = operatorHolderRefToBondID[operator][holder][referenceID];
        return bondID;
    }

    function balanceOf(address account) public view returns (uint256) {
        return account.balance;
    }

    function init_state() public {}

    function getDelegatedAuthority(address a) public view returns (address) {
        return delegatedAuthority[a];
    }

    function deposit(address operator) public payable {
        require(msg.sender != address(0));
        everDeposited[operator] = everDeposited[operator].add(msg.value);
        super.deposit(operator);
    }

    /* to track totalLockedBond we add a map (operator, holder, referenceID) to BondID */

    function registerBondID(
        address operator,
        address holder,
        uint256 referenceID
    ) private {
        /*  bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );
        operatorHolderRefToBondID[operator][holder][referenceID] = bondID;
        */
    }

    function createBond(
        address operator,
        address holder,
        uint256 referenceID,
        uint256 amount,
        address authorizedSortitionPool
    ) public {
        require(msg.sender != address(0));
        registerBondID(operator, holder, referenceID);
        super.createBond(
            operator,
            holder,
            referenceID,
            amount,
            authorizedSortitionPool
        );
    }

    function reassignBond(
        address operator,
        uint256 referenceID,
        address newHolder,
        uint256 newReferenceID
    ) public {
        require(msg.sender != address(0));
        registerBondID(operator, msg.sender, referenceID);
        registerBondID(operator, newHolder, newReferenceID);
        super.reassignBond(operator, referenceID, newHolder, newReferenceID);
    }

    function freeBond(address operator, uint256 referenceID) public {
        require(msg.sender != address(0));
        registerBondID(operator, msg.sender, referenceID);
        super.freeBond(operator, referenceID);
    }

    function seizeBond(
        address operator,
        uint256 referenceID,
        uint256 amount,
        address payable destination
    ) public {
        require(msg.sender != address(0));
        // so transfer will not fail
        require(destination == otherBeneficiary);
        registerBondID(operator, msg.sender, referenceID);
        super.seizeBond(operator, referenceID, amount, otherBeneficiary);
    }

    function _hasDelegationLockPassed(address operator) public returns (bool) {
        return hasDelegationLockPassed(operator);
    }
}
