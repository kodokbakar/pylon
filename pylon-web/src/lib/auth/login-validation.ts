export type LoginFormData = {
  email: string;
  password: string;
};

export type LoginFormErrors = Partial<Record<keyof LoginFormData, string>>;

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function validateLoginForm(formData: LoginFormData): LoginFormErrors {
  const errors: LoginFormErrors = {};
  const email = formData.email.trim();

  if (email === "") {
    errors.email = "Email is required.";
  } else if (!emailPattern.test(email)) {
    errors.email = "Enter a valid email address.";
  }

  if (formData.password.length === 0) {
    errors.password = "Password is required.";
  }

  return errors;
}

export function hasLoginFormErrors(errors: LoginFormErrors): boolean {
  return Object.values(errors).some(Boolean);
}
