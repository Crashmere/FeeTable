export function calculate(quantity: string, price: string): string | null {
  quantity = quantity.trim();
  price = price.trim();
  if (
    !/^-?[0-9]+(?:[.][0-9]{1,3})?$/.test(quantity) ||
    !/^[0-9]+(?:[.][0-9]{1,2})?$/.test(price) ||
    quantity.length > 24 ||
    price.length > 24
  )
    return null;
  const negative = quantity.startsWith("-");
  const [whole, fraction = ""] = quantity.replace("-", "").split(".");
  const q = BigInt(whole) * 1000n + BigInt(fraction.padEnd(3, "0"));
  const [priceWhole, priceFraction = ""] = price.split(".");
  const p = BigInt(priceWhole) * 100n + BigInt(priceFraction.padEnd(2, "0"));
  const cents = (q * p + 500n) / 1000n;
  const max = 1_000_000_000_000n;
  if (q > max || p > max || cents > max) return null;
  return (
    (negative && cents !== 0n ? "-" : "") +
    cents / 100n +
    "." +
    (cents % 100n).toString().padStart(2, "0")
  );
}
export function money(value: string) {
  const [whole, decimal = "00"] = value.split(".");
  const negative = whole.startsWith("-");
  const digits = negative ? whole.slice(1) : whole;
  const groups: string[] = [];
  for (let end = digits.length; end > 0; end -= 3)
    groups.unshift(digits.slice(Math.max(0, end - 3), end));
  return (negative ? "-" : "") + groups.join(",") + "." + decimal;
}
