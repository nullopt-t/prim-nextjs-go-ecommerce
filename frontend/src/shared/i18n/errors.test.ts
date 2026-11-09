import { describe, expect, it } from "vitest";
import { getErrorMessageByCode } from "./errors";
import { apiErrorMessage } from "../lib/notify";
import { ApiError } from "../api/client";

describe("Arabized backend error code mapping", () => {
  it("translates canonical error codes in English", () => {
    expect(getErrorMessageByCode("INSUFFICIENT_INVENTORY", "en")).toBe(
      "Insufficient inventory available for this item."
    );
    expect(getErrorMessageByCode("INVALID_ORDER_STATUS", "en")).toBe(
      "The requested order status transition is invalid."
    );
    expect(getErrorMessageByCode("COUPON_NOT_FOUND", "en")).toBe(
      "The specified coupon code was not found or is expired."
    );
    expect(getErrorMessageByCode("UNAUTHORIZED", "en")).toBe(
      "You are not authorized to perform this action. Please sign in."
    );
    expect(getErrorMessageByCode("INVALID_CREDENTIALS", "en")).toBe(
      "The email or password you entered is incorrect."
    );
  });

  it("translates canonical error codes in Arabic", () => {
    expect(getErrorMessageByCode("INSUFFICIENT_INVENTORY", "ar")).toBe(
      "الكمية المتوفرة في المخزون غير كافية لهذا المنتج."
    );
    expect(getErrorMessageByCode("INVALID_ORDER_STATUS", "ar")).toBe(
      "حالة الطلب المطلوبة غير صالحة."
    );
    expect(getErrorMessageByCode("COUPON_NOT_FOUND", "ar")).toBe(
      "رمز القسيمة غير موجود أو منتهي الصلاحية."
    );
    expect(getErrorMessageByCode("UNAUTHORIZED", "ar")).toBe(
      "غير مصرح لك بالقيام بهذا الإجراء. يرجى تسجيل الدخول."
    );
    expect(getErrorMessageByCode("INVALID_CREDENTIALS", "ar")).toBe(
      "البريد الإلكتروني أو كلمة المرور غير صحيحة."
    );
  });

  it("apiErrorMessage prioritizes translated error code over raw backend strings in Arabic", () => {
    const apiErr = new ApiError("raw backend error message", 400, "INSUFFICIENT_INVENTORY");
    const result = apiErrorMessage(apiErr, "Fallback message", "ar");
    expect(result).toBe("الكمية المتوفرة في المخزون غير كافية لهذا المنتج.");
  });

  it("apiErrorMessage handles plain object with error code", () => {
    const res = { code: "INVALID_CREDENTIALS", error: "bad password" };
    expect(apiErrorMessage(res, "Fallback", "ar")).toBe("البريد الإلكتروني أو كلمة المرور غير صحيحة.");
    expect(apiErrorMessage(res, "Fallback", "en")).toBe("The email or password you entered is incorrect.");
  });

  it("apiErrorMessage falls back to raw backend message for unknown error codes", () => {
    const apiErr = new ApiError("Custom backend failure", 400, "UNKNOWN_CUSTOM_CODE");
    expect(apiErrorMessage(apiErr, "Fallback", "ar")).toBe("Custom backend failure");
  });

  it("apiErrorMessage falls back to fallback string when no message is present", () => {
    expect(apiErrorMessage({}, "Default fallback", "ar")).toBe("Default fallback");
    expect(apiErrorMessage(null, "Default fallback", "en")).toBe("Default fallback");
  });
});
