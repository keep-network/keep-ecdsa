pragma solidity ^0.5.4;

import "../../contracts/ECDSAKeepFactory.sol";

/// @title ECDSA Keep Factory Stub
/// @dev This contract is for testing purposes only.
contract ECDSAKeepFactoryStub is ECDSAKeepFactory {

    // @dev Returns list of registered members.
    function getMemberCandidates() public view returns (address payable[] memory){
        return memberCandidates;
    }

    // @dev Returns calculated keep address.
    function openKeepStub(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) public payable returns (address) {
        _groupSize;
        _honestThreshold;
        _owner;

        return calculateKeepAddress();
    }

    /// @dev Calculates an address for a keep based on the address of the factory.
    /// We need it to have predictable addresses for factories verification.
    function calculateKeepAddress() public view returns (address) {
        uint256 factoryAddressInt = uint256(address(this));
        return address(factoryAddressInt % 1000000000000);
    }

    /**
     * @dev Gets the number of submitted group candidate tickets so far.
     */
    function submittedTicketsCount() public view returns (uint256) {
        return tickets.length;
    }

    function getGroupSelectionRelayEntry() public view returns (uint256) {
        return seed;
    }

}
