import BigNumber from "bignumber.js"

const decimalPlaces = 2
const noDecimalPlaces = 0
const format = {
  groupSeparator: "",
  decimalSeparator: ".",
}

export const shorten18Decimals = (value) =>
  toFormat(new BigNumber(value).dividedBy(new BigNumber(1e18)))

export const toFormat = (value, decimals = true, rounding) =>
  new BigNumber(value).toFormat(
    decimals ? decimalPlaces : noDecimalPlaces,
    rounding,
    format
  )
