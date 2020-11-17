certoraRun.py specs/harness/FullyBackedBondingHarness.sol:FullyBackedBondingHarness  \
specs/harness/Beneficiary.sol:Beneficiary specs/harness/OtherBeneficiary.sol:OtherBeneficiary \
--structLink FullyBackedBondingHarness:2=Beneficiary \
--link FullyBackedBondingHarness:otherBeneficiary=OtherBeneficiary \
--solc solc5.17  --verify FullyBackedBondingHarness:specs/fullyBackedBonding.spec \
--settings -recursionErrorAsAssert=false,-deleteSMTFile=true,-graphDrawLimit=0,-enableStorageAnalysis=true \
--staging --cache FullyBackedBonding 
