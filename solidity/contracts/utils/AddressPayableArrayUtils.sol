pragma solidity ^0.5.4;

// TODO: This is library copied from keep-core `AddressArrayUtils `with the
// modification of address type to payable. When merging repositories we need
// to combine the utils.
library AddressPayableArrayUtils {
    function contains(address payable[] memory self, address _address)
        internal
        pure
        returns (bool)
    {
        for (uint256 i = 0; i < self.length; i++) {
            if (_address == self[i]) {
                return true;
            }
        }
        return false;
    }
}
