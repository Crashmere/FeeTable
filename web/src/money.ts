export function calculate(quantity: string, price: string): string | null {
  if (
    !/^-?[0-9]+(?:[.][0-9]{1,3})?$/.test(quantity) ||
    !/^[0-9]+$/.test(price) ||
    quantity.length > 20 ||
    price.length > 12
  )
    return null;
  const negative = quantity.startsWith("-");
  const [whole, fraction = ""] = quantity.replace("-", "").split(".");
  const q = BigInt(whole) * 1000n + BigInt(fraction.padEnd(3, "0"));
  const cents = (q * BigInt(price) + 5n) / 10n;
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
