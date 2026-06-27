import { describe, expect, it } from "vitest";

import {
  hasLoginFormErrors,
  validateLoginForm,
  type LoginFormData,
} from "./login-validation";

const validForm: LoginFormData = {
  email: "alice@example.com",
  password: "password123",
};

describe("login validation", () => {
  it("accepts valid login data", () => {
    const errors = validateLoginForm(validForm);

    expect(errors).toEqual({});
    expect(hasLoginFormErrors(errors)).toBe(false);
  });

  it("validates email format", () => {
    const errors = validateLoginForm({
      ...validForm,
      email: "invalid-email",
    });

    expect(errors.email).toBe("Enter a valid email address.");
    expect(hasLoginFormErrors(errors)).toBe(true);
  });

  it("validates required password", () => {
    const errors = validateLoginForm({
      ...validForm,
      password: "",
    });

    expect(errors.password).toBe("Password is required.");
    expect(hasLoginFormErrors(errors)).toBe(true);
  });
});
