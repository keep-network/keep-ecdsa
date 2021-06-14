module.exports = {
  ...require("@keep-network/prettier-config-keep"),
  plugins: ["prettier-plugin-sh", "prettier-plugin-toml"],
  overrides: [
    {
      files: "*.toml.SAMPLE",
      options: {
        parser: "toml",
      },
    },
  ],
}
