import { describe, it, expect } from "vitest";
import { calculate, money } from "./money";
describe("decimal preview", () => {
  it("rounds exact decimal products", () => {
    expect(calculate("0.145", "1")).toBe("0.15");
    expect(calculate("-1.025", "1")).toBe("-1.03");
    expect(calculate("2.125", "40")).toBe("85.00");
    expect(calculate("3", "0")).toBe("0.00");
    expect(calculate("10.5", "2.88")).toBe("30.24");
    expect(calculate("0.5", "0.01")).toBe("0.01");
    expect(calculate("-0.5", "0.01")).toBe("-0.01");
    expect(calculate(" 10.5 ", " 2.88 ")).toBe("30.24");
  });
  it("rejects invalid input without throwing", () => {
    for (const value of ["", "NaN", "1e2", "1,25", "1x25", "0.1234"])
      expect(calculate(value, "1")).toBeNull();
    for (const price of ["2.888", "-1", "1e2", "10000000000.01"])
      expect(calculate("10.5", price)).toBeNull();
  });
  it("groups exact strings", () => {
    expect(money("1234567.89")).toBe("1,234,567.89");
    expect(money("-1234.50")).toBe("-1,234.50");
  });
});
