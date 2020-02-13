pragma solidity ^0.5.4;

import "@keep-network/keep-core/contracts/Registry.sol";

interface TokenStaking {
    function isAuthorizedForOperator(address operator, address operatorContract) external view returns (bool);
    function authorizerOf(address operator) external view returns (address);
}

// TODO: This contract is expected to implement functions defined by IBonding
// interface defined in @keep-network/sortition-pools. After merging the
// repositories we need to move IBonding definition to sit closer to KeepBonding
// contract so that sortition pools import it for own needs. It is the bonding
// module which should define an interface, and sortition pool module should be
// just importing it.

/// @title Keep Bonding
/// @notice Contract holding deposits from keeps' operators.
// TODO: Update KeepBonding contract implementation to new requirements based
// on the spec.
contract KeepBonding {
    // Registry contract with a list of approved factories (operator contracts) and upgraders.
    Registry internal registry;

    // Staking contract linked to this contract.
    TokenStaking internal stakingContract;

    // Unassigned ether values deposited by operators.
    mapping(address => uint256) internal unbondedValue;
    // References to created bonds. Bond identifier is built from operator's
    // address, holder's address and reference ID assigned on bond creation.
    mapping(bytes32 => uint256) internal lockedBonds;

    // Sortition pools authorized by operator authorizer.
    mapping(address => mapping (address => bool)) internal authorizedPools;

    /// @dev Initialize Keep Bonding contract.
    /// @param registryAddress Keep registry contract linked to this contract.
    /// @param stakingContractAddress Keep Token staking contract linked to this contract.
    constructor(address registryAddress, address stakingContractAddress) public {
        registry = Registry(registryAddress);
        stakingContract = TokenStaking(stakingContractAddress);
    }

    /// @notice Returns the amount of ether the operator has made available for
    /// bonding by the bond creator. If the operator doesn't exists or the operator
    /// doesn't exists returns zero.
    /// @dev Implements function expected by sortition pools' IBonding interface.
    /// @param operator Address of the operator.
    /// @param bondCreator Address authorized to create a bond.
    /// @param authorizedSortitionPool Address of authorized sortition pool.
    /// @return Amount of deposited ether available for bonding.
    function availableUnbondedValue(
        address operator,
        address bondCreator,
        address authorizedSortitionPool
    ) public view returns (uint256) {
        if (registry.isApprovedOperatorContract(bondCreator) &&
            stakingContract.isAuthorizedForOperator(bondCreator, operator) &&
            hasSecondaryAuthorization(operator, authorizedSortitionPool)
        ) {
            return unbondedValue[operator];
        } else {
            return 0;
        }
    }

    /// @notice Returns value of ether available for bonding for the operator.
    /// @param operator Address of the operator.
    /// @return Value of deposited ether available for bonding.
    /// TODO: Should be replaced by availableUnbondedValue.
    function availableBondingValue(address operator)
        public
        view
        returns (uint256)
    {
        return unbondedValue[operator];
    }

    /// @notice Add ether to operator's value available for bonding.
    /// @param operator Address of the operator.
    function deposit(address operator) external payable {
        unbondedValue[operator] += msg.value;
    }

    /// @notice Draw amount from sender's value available for bonding.
    /// @param amount Value to withdraw.
    /// @param destination Address to send the amount to.
    function withdraw(uint256 amount, address payable destination) public {
        require(
            availableBondingValue(msg.sender) >= amount,
            "Insufficient unbonded value"
        );

        unbondedValue[msg.sender] -= amount;

        (bool success, ) = destination.call.value(amount)("");
        require(success, "Transfer failed");
    }

    /// @notice Create bond for given operator, holder, reference and amount.
    /// @dev Function can be executed only by authorized contract. Reference ID
    /// should be unique for holder and operator.
    /// @param operator Address of the operator to bond.
    /// @param holder Address of the holder of the bond.
    /// @param referenceID Reference ID used to track the bond by holder.
    /// @param amount Value to bond.
    function createBond(
        address operator,
        address holder,
        uint256 referenceID,
        uint256 amount
    ) public onlyAuthorized {
        require(
            // TODO: Provide authorized sortition pool address
            availableUnbondedValue(operator, msg.sender, address(0)) >= amount,
            "Insufficient unbonded value"
        );

        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        require(
            lockedBonds[bondID] == 0,
            "Reference ID not unique for holder and operator"
        );

        unbondedValue[operator] -= amount;
        lockedBonds[bondID] += amount;
    }

    /// @notice Returns value of ether bonded for the operator.
    /// @param operator Address of the operator.
    /// @param holder Address of the holder of the bond.
    /// @param referenceID Reference ID used to track the bond by holder.
    /// @return Operator's bonded ether.
    function bondAmount(address operator, address holder, uint256 referenceID)
        public
        view
        returns (uint256)
    {
        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        return lockedBonds[bondID];
    }

    /// @notice Reassigns a bond to a new holder under a new reference.
    /// @dev Function requires that a caller is the holder of the bond which is
    /// being reassigned.
    /// @param operator Address of the bonded operator.
    /// @param referenceID Reference ID of the bond.
    /// @param newHolder Address of the new holder of the bond.
    /// @param newReferenceID New reference ID to register the bond.
    function reassignBond(
        address operator,
        uint256 referenceID,
        address newHolder,
        uint256 newReferenceID
    ) public {
        address holder = msg.sender;
        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        require(lockedBonds[bondID] > 0, "Bond not found");

        bytes32 newBondID = keccak256(
            abi.encodePacked(operator, newHolder, newReferenceID)
        );

        require(
            lockedBonds[newBondID] == 0,
            "Reference ID not unique for holder and operator"
        );

        lockedBonds[newBondID] = lockedBonds[bondID];
        lockedBonds[bondID] = 0;
    }

    /// @notice Releases the bond and moves the bond value to the operator's
    /// unbounded value pool.
    /// @dev Function requires that a caller is the holder of the bond which is
    /// being released.
    /// @param operator Address of the bonded operator.
    /// @param referenceID Reference ID of the bond.
    function freeBond(address operator, uint256 referenceID) public {
        address holder = msg.sender;
        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        require(lockedBonds[bondID] > 0, "Bond not found");

        uint256 amount = lockedBonds[bondID];
        lockedBonds[bondID] = 0;
        unbondedValue[operator] = amount;
    }

    /// @notice Seizes the bond by moving some or all of a locked bond to holder's
    /// account.
    /// @dev Function requires that a caller is the holder of the bond which is
    /// being seized.
    /// @param operator Address of the bonded operator.
    /// @param referenceID Reference ID of the bond.
    /// @param amount Amount to be seized.
    /// @param destination Address to send the amount to.
    function seizeBond(
        address operator,
        uint256 referenceID,
        uint256 amount,
        address payable destination
    ) public {
        require(amount > 0, "Requested amount should be greater than zero");

        address payable holder = msg.sender;
        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        require(
            lockedBonds[bondID] >= amount,
            "Requested amount is greater than the bond"
        );

        lockedBonds[bondID] -= amount;

        (bool success, ) = destination.call.value(amount)("");
        require(success, "Transfer failed");
    }

    /// @notice Checks if the caller is an authorized contract.
    /// @dev Throws an error if called by any account other than one of the authorized
    /// contracts.
    modifier onlyAuthorized() {
        // TODO: Add authorization checks.
        require(
            registry.isApprovedOperatorContract(msg.sender),
            "Factory contract is not approved"
        );
        _;
    }


    /// @dev Authorizes sortition pool for provided operator
    /// @dev Only operator authorizer can call this function
    function authorizeSortitionPoolContract(address _operator, address _poolAddress) public {
        require(
            stakingContract.authorizerOf(_operator) == msg.sender,
            "Not authorized"
        );
        authorizedPools[msg.sender][_poolAddress] = true;
    }

    /// @notice Checks if the sortition pool has been authorized for provided operator by its authorizer.
    function hasSecondaryAuthorization(address _operator, address _poolAddress) public view returns (bool) {
        address authorizer = stakingContract.authorizerOf(_operator);
        return authorizedPools[authorizer][_poolAddress];
    }
}
