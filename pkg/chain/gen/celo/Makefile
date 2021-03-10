# Environment provides the solidity directory as a potentially-relative path,
# which we resolve. Then we resolve the Solidity files in a contracts/ directory
# at that path.
solidity_dir=$(realpath ${SOLIDITY_DIR})
solidity_files := $(wildcard ${solidity_dir}/contracts/*.sol)

# Bare Solidity filenames without .sol or Solidity directory prefix.
contract_stems := $(notdir $(basename $(solidity_files)))
# *ImplV1.go files will get generated into clean Keep contract bindings, the
# corresponding contract filenames will drop the ImplV1, if it exists, and live
# in the contract/ directory.
clean_contract_stems := $(filter %ImplV1,$(contract_stems)) $(filter BondedECDSAKeepFactory, $(contract_stems)) $(filter BondedECDSAKeep, $(contract_stems))
contract_files := $(addprefix contract/,$(addsuffix .go,$(subst ImplV1,,$(clean_contract_stems))))
# Go abigen bindings in abi/ subdirectory with .go suffix, alongside solc ABI
# files with .abi suffix.
abi_files := $(addprefix abi/,$(addsuffix .abi,$(clean_contract_stems)))
abigen_files := $(addprefix abi/,$(addsuffix .go,$(clean_contract_stems)))

# Additional build tags which should be passed while running `abigen` command.
# The `default` value of the `ABIGEN_BUILD_TAGS` env variable is an arbitrary
# default value and doesn't have any special meaning. Without it, the `go run`
# command which runs the abigen will fail due to empty build tag list in case
# `ABIGEN_BUILD_TAGS` env variable is not set.
ABIGEN_BUILD_TAGS ?= default
abigen_build_tags = ${ABIGEN_BUILD_TAGS}

all: gen_contract_go gen_abi_go

clean:
	rm -r abi/*
	rm -r contract/*
	mkdir tmp && mv cmd/cmd*.go tmp
	rm -r cmd/*
	mv tmp/* cmd && rm -r tmp

gen_abi_go: $(abigen_files)

gen_contract_go: $(contract_files)

abi/%.abi: ${solidity_dir}/contracts/%.sol
	solc solidity-bytes-utils/=${solidity_dir}/node_modules/solidity-bytes-utils/ \
		 openzeppelin-solidity/=${solidity_dir}/node_modules/openzeppelin-solidity/ \
		 @openzeppelin/upgrades/=${solidity_dir}/node_modules/@openzeppelin/upgrades/ \
		 @keep-network/keep-core/=${solidity_dir}/node_modules/@keep-network/keep-core/  \
		 @keep-network/sortition-pools/=${solidity_dir}/node_modules/@keep-network/sortition-pools/  \
		 --allow-paths ${solidity_dir} \
		 --overwrite \
		 --abi \
		 -o abi $<


abi/%.go: abi/%.abi
	go run -tags ${abigen_build_tags} github.com/celo-org/celo-blockchain/cmd/abigen --abi $< --pkg abi --type $* --out $@

contract/%.go cmd/%.go: abi/%ImplV1.abi abi/%ImplV1.go abi/%.go *.go
	go run github.com/keep-network/keep-common/tools/generators/ethlike \
		-host-chain-module github.com/celo-org/celo-blockchain \
		-chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil \
		-config-func config.ReadCeloConfig \
		$< contract/$*.go cmd/$*.go

contract/BondedECDSAKeepFactory.go cmd/BondedECDSAKeepFactory.go: abi/BondedECDSAKeepFactory.abi abi/BondedECDSAKeepFactory.go *.go
	go run github.com/keep-network/keep-common/tools/generators/ethlike \
		-host-chain-module github.com/celo-org/celo-blockchain \
		-chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil \
		-config-func config.ReadCeloConfig \
		$< contract/BondedECDSAKeepFactory.go cmd/BondedECDSAKeepFactory.go \

contract/BondedECDSAKeep.go cmd/BondedECDSAKeep.go: abi/BondedECDSAKeep.abi abi/BondedECDSAKeep.go *.go
	go run github.com/keep-network/keep-common/tools/generators/ethlike \
		-host-chain-module github.com/celo-org/celo-blockchain \
		-chain-util-package github.com/keep-network/keep-common/pkg/chain/celo/celoutil \
		-config-func config.ReadCeloConfig \
		$< contract/BondedECDSAKeep.go cmd/BondedECDSAKeep.go \