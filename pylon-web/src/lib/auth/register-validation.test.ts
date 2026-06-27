import { describe, expect, it } from "vitest";

import {
  hasRegisterFormErrors,
  validateRegisterForm,
  type RegisterFormData,
} from "./register-validation";

const validForm: RegisterFormData = {
  username: "alice123",
  email: "alice@example.com",
  password: "password123",
  confirmPassword: "password123",
};

describe("register validation", () => {
  it("accepts valid register data", () => {
    const errors = validateRegisterForm(validForm);

    expect(errors).toEqual({});
    expect(hasRegisterFormErrors(errors)).toBe(false);
  });

  it("validates username, email, and password rules", () => {
    const errors = validateRegisterForm({
      username: "a!",
      email: "invalid-email",
      password: "short",
      confirmPassword: "short",
    });

    expect(errors).toMatchObject({
      username: "Username must be at least 3 characters.",
      email: "Enter a valid email address.",
      password: "Password must be at least 8 characters.",
    });
    expect(hasRegisterFormErrors(errors)).toBe(true);
  });

  it("validates password confirmation", () => {
    const errors = validateRegisterForm({
      ...validForm,
      confirmPassword: "different123",
    });

    expect(errors.confirmPassword).toBe("Passwords do not match.");
    expect(hasRegisterFormErrors(errors)).toBe(true);
  });
});
