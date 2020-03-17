pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

// TODO: This contract is expected to implement functions defined by IBonding
// interface defined in @keep-network/sortition-pools. After merging the
// repositories we need to move IBonding definition to sit closer to KeepBonding
// contract so that sortition pools import it for own needs. It is the bonding
// module which should define an interface, and sortition pool module should be
// just importing it.

/// @title Abstract Bonding
/// @notice Shared code for holding deposits from keeps' operators.
contract AbstractBonding {
    using SafeMath for uint256;

    // Unassigned value in wei deposited by operators.
    mapping(address => uint256) public unbondedValue;

    // References to created bonds. Bond identifier is built from operator's
    // address, holder's address and reference ID assigned on bond creation.
    mapping(bytes32 => uint256) internal lockedBonds;

    // Sortition pools authorized by operator's authorizer.
    // operator -> pool -> boolean
    mapping(address => mapping (address => bool)) internal authorizedPools;

    /// @notice Returns the amount of wei the operator has made available for
    /// bonding and that is still unbounded. If the operator doesn't exist or
    /// bond creator is not authorized as an operator contract or it is not
    /// authorized by the operator or there is no secondary authorization for
    /// the provided sortition pool, function returns 0.
    /// @dev Implements function expected by sortition pools' IBonding interface.
    /// @param operator Address of the operator.
    /// @param bondCreator Address authorized to create a bond.
    /// @param authorizedSortitionPool Address of authorized sortition pool.
    /// @return Amount of authorized wei deposit available for bonding.
    function availableUnbondedValue(
        address operator,
        address bondCreator,
        address authorizedSortitionPool
    ) public view returns (uint256);

    /// @notice Add the provided value to operator's pool available for bonding.
    /// @param operator Address of the operator.
    function deposit(address operator) external payable;

    /// @notice Withdraws amount from operator's value available for bonding.
    /// Can be called only by the operator or by the stake owner.
    /// @param amount Value to withdraw in wei.
    /// @param operator Address of the operator.
    function withdraw(uint256 amount, address operator) public {
        require(
            _isWithdrawerOf(msg.sender, operator),
            "Not authorized to withdraw bond"
        );

        require(
            unbondedValue[operator] >= amount,
            "Insufficient unbonded value"
        );

        unbondedValue[operator] = unbondedValue[operator].sub(amount);

        (bool success, ) = _beneficiaryOf(operator).call.value(amount)("");
        require(success, "Transfer failed");
    }

    /// @notice Create bond for the given operator, holder, reference and amount.
    /// @dev Function can be executed only by authorized contract. Reference ID
    /// should be unique for holder and operator.
    /// @param operator Address of the operator to bond.
    /// @param holder Address of the holder of the bond.
    /// @param referenceID Reference ID used to track the bond by holder.
    /// @param amount Value to bond in wei.
    /// @param authorizedSortitionPool Address of authorized sortition pool.
    function createBond(
        address operator,
        address holder,
        uint256 referenceID,
        uint256 amount,
        address authorizedSortitionPool
    ) public {
        require(
            availableUnbondedValue(
                operator,
                msg.sender,
                authorizedSortitionPool
            ) >= amount,
            "Insufficient unbonded value"
        );

        bytes32 bondID = keccak256(
            abi.encodePacked(operator, holder, referenceID)
        );

        require(
            lockedBonds[bondID] == 0,
            "Reference ID not unique for holder and operator"
        );

        unbondedValue[operator] = unbondedValue[operator].sub(amount);
        lockedBonds[bondID] = lockedBonds[bondID].add(amount);
    }

    /// @notice Returns value of wei bonded for the operator.
    /// @param operator Address of the operator.
    /// @param holder Address of the holder of the bond.
    /// @param referenceID Reference ID of the bond.
    /// @return Amount of wei in the selected bond.
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
    /// @dev Function requires that a caller is the current holder of the bond
    /// which is being reassigned.
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
    /// @dev Function requires that caller is the holder of the bond which is
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
        unbondedValue[operator] = unbondedValue[operator].add(amount);
    }

    /// @notice Seizes the bond by moving some or all of the locked bond to the
    /// provided destination address.
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

        lockedBonds[bondID] = lockedBonds[bondID].sub(amount);

        (bool success, ) = destination.call.value(amount)("");
        require(success, "Transfer failed");
    }

    /// @notice Authorizes sortition pool for the provided operator.
    /// Operator's authorizers need to authorize individual sortition pools
    /// per application since they may be interested in participating only in
    /// a subset of keep types used by the given appliction.
    /// @dev Only operator's authorizer can call this function.
    function authorizeSortitionPoolContract(
        address _operator,
        address _poolAddress
    ) public {
        require(
            _isAuthorizerOf(msg.sender, _operator),
            "Not authorized"
        );
        authorizedPools[_operator][_poolAddress] = true;
    }

    /// @notice Checks if the sortition pool has been authorized for the
    /// provided operator by its authorizer.
    /// @dev See authorizeSortitionPoolContract.
    function hasSecondaryAuthorization(
        address _operator,
        address _poolAddress
    ) public view returns (bool) {
        return authorizedPools[_operator][_poolAddress];
    }

    /// @notice Return whether _sender may withdraw for _operator.
    function _isWithdrawerOf(
        address _sender,
        address _operator
    ) internal view returns (bool);

    /// @notice Return whether _sender is the authorizer of _operator.
    function _isAuthorizerOf(
        address _sender,
        address _operator
    ) internal view returns (bool);

    /// @notice Get the beneficiary of _operator.
    function _beneficiaryOf(
        address _operator
    ) internal view returns (address payable);
}
