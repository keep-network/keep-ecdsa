const testValues = {
  interval0: {
    amountToAllocate: 128,
    merkleRoot:
      "0x65b315f4565a40f738cbaaef7dbab4ddefa14620407507d0f2d5cdbd1d8063f6",
    claims: [
      {
        index: "0",
        amount: "85",
        operator: "0x012ed55a0876Ea9e58277197DC14CbA47571CE28",
        proof: [
          "0x1419d8cdecc66122fdd450e7322c82dbd1a183b4abad7ed01637042fcbbb1231",
          "0x79be08e672b91905bbbfa785ba28cf962112aae1bc30911e74e1af3939e7501f",
          "0x0a9ae0952dd27639b419a19a4915049dab8d969025dd2b5dd3d41d189ab741b6",
          "0xff2cdff30a8b335d489ac2b79ae2807547ab391cd83f208efe09383cd43c6a1b",
          "0x37a9403fe61c38ad4a12b00106d091994901f4a6365af9584ff4b2710acb4961",
          "0x023527b3cb4eb23b75f8554373ef468c6cc5a446a5bbf5b26133d684a82dc8ee",
          "0xa8390642d0b4fcbcfd25bd2787f9e498ce1f24c3d630f57a1560354a5b4dd06e",
        ],
        beneficiary: "0x82Eda22AE1d4B16C11261F770C07092b0A62136a",
      },
      {
        index: "6",
        amount: "42",
        operator: "0x162F49fE6F365d04Db07F77377699aeFE2E8A2cf",
        proof: [
          "0x098f41b175b4374631df722f58f4bc63fca57851682d6dcc911a975f4582f385",
          "0xf4b6ad879c695dbdb0376fca8b9548cc4718bd08e0e657732b3b9df595a71dba",
          "0xc37bf5b48fd7a4640107ef123ae21c5af58e79266cf65e41bbbed0480c3d2c08",
          "0xedc664bfdaba16701ff14c0fbd5c2cd9e349f6d9051f837b6848d361c7c0b59c",
          "0x37a9403fe61c38ad4a12b00106d091994901f4a6365af9584ff4b2710acb4961",
          "0x023527b3cb4eb23b75f8554373ef468c6cc5a446a5bbf5b26133d684a82dc8ee",
          "0xa8390642d0b4fcbcfd25bd2787f9e498ce1f24c3d630f57a1560354a5b4dd06e",
        ],
        beneficiary: "0x6ab392c2134a6cE826cFBb0045175fAcCF504F69",
      },
      {
        index: "77",
        amount: "1",
        operator: "0xF3c6F5F265F503f53EAD8aae90FC257A5aa49AC1",
        proof: [
          "0x081122f7920ab9553b71a84d1f9961256379ed7ff5257c8bd9749155e3dddeff",
          "0xb335096692ef570690f2d858f2d52c268728d60b12a2a856f2691155ccf36377",
          "0xc0663d801b44d93fde14c740757de07a09feb3d8c092d75315a720dc28d10caa",
          "0xedc664bfdaba16701ff14c0fbd5c2cd9e349f6d9051f837b6848d361c7c0b59c",
          "0x37a9403fe61c38ad4a12b00106d091994901f4a6365af9584ff4b2710acb4961",
          "0x023527b3cb4eb23b75f8554373ef468c6cc5a446a5bbf5b26133d684a82dc8ee",
          "0xa8390642d0b4fcbcfd25bd2787f9e498ce1f24c3d630f57a1560354a5b4dd06e",
        ],
        beneficiary: "0xfc8437956EeaCCE97996Ad01af59b828F3F5A808",
      },
    ],
  },
  interval1: {
    amountToAllocate: 160,
    merkleRoot:
      "0x226999926d79bc999598889f80858deefe48c9570cebdd887fe038947c67313a",
    claims: [
      {
        index: "0",
        amount: "80",
        operator: "0x05fc93DeFFFe436822100E795F376228470FB514",
        proof: [
          "0x54fce5fe1df9254fa00885de1ee9455b85f7dadc4b3192c33e5c3be9bea8d060",
          "0xeddbf7ea36eda7b37bd9314041a68b1c8adc2d5351ee1b373ab5fc464d20b02a",
          "0x162ab1201ce6d116732a2a8793945a23244afb8f79242255d3c34fea6aeceb73",
          "0x172864ef2714d9feaf400fe1342b595b61fc740234de3a7cd4e3d4eb8be3fc36",
        ],
        beneficiary: "0x5Ab6DE7b08baF7A5d6Cb475b48E9055685E5C346",
      },
      {
        index: "1",
        amount: "70",
        operator: "0x57E7c6B647C004CFB7A38E08fDDef09Af5Ea55eD",
        proof: [
          "0xc5d54b41f2c330a898fa1773bc4e8355b9bcc6e382c055f549b845af3814790c",
          "0x303ed1284f356d779e4a0a48644f5c4a210f7ea3f250d14ad811767cfe059a39",
          "0x20779de3e5a9a44f012cedf6f56e78af493fcb441d19cd884471e28457bbf697",
          "0x172864ef2714d9feaf400fe1342b595b61fc740234de3a7cd4e3d4eb8be3fc36",
        ],
        beneficiary: "0x5664bc69ABeb9CB25c0CAA0C9326C8217E8AF6B4",
      },
      {
        index: "7",
        amount: "10",
        operator: "0xF3c6F5F265F503f53EAD8aae90FC257A5aa49AC1",
        proof: [
          "0x92a4eaf5cccec93f36a5e27edf10b095949c5517f6a4896295f18b34652fbf03",
          "0xd7c0e1ce86eac4aec04c6f30dbd9d9215d1d30a640cd92c09cbcf68648711ea0",
          "0x20779de3e5a9a44f012cedf6f56e78af493fcb441d19cd884471e28457bbf697",
          "0x172864ef2714d9feaf400fe1342b595b61fc740234de3a7cd4e3d4eb8be3fc36",
        ],
        beneficiary: "0xfc8437956EeaCCE97996Ad01af59b828F3F5A808",
      },
    ],
  },
}

module.exports = {testValues}
