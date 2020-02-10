pragma solidity ^0.5.4;

library AddressArrayUtils {
    function contains(address payable[] memory self, address _address)
        internal
        pure
        returns (bool)
    {
        for (uint i = 0; i < self.length; i++) {
            if (_address == self[i]) {
                return true;
            }
        }
        return false;
    }
}
