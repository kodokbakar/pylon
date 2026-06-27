export type RegisterFormData = {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export type RegisterFormErrors = Partial<
  Record<keyof RegisterFormData, string>
>;

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const usernamePattern = /^[a-zA-Z0-9]+$/;

export function validateRegisterForm(
  formData: RegisterFormData,
): RegisterFormErrors {
  const errors: RegisterFormErrors = {};
  const username = formData.username.trim();
  const email = formData.email.trim();

  if (username === "") {
    errors.username = "Username is required.";
  } else if (username.length < 3) {
    errors.username = "Username must be at least 3 characters.";
  } else if (!usernamePattern.test(username)) {
    errors.username = "Username can only contain letters and numbers.";
  }

  if (email === "") {
    errors.email = "Email is required.";
  } else if (!emailPattern.test(email)) {
    errors.email = "Enter a valid email address.";
  }

  if (formData.password.length === 0) {
    errors.password = "Password is required.";
  } else if (formData.password.length < 8) {
    errors.password = "Password must be at least 8 characters.";
  }

  if (formData.confirmPassword.length === 0) {
    errors.confirmPassword = "Confirm your password.";
  } else if (formData.confirmPassword !== formData.password) {
    errors.confirmPassword = "Passwords do not match.";
  }

  return errors;
}

export function hasRegisterFormErrors(errors: RegisterFormErrors): boolean {
  return Object.values(errors).some(Boolean);
}
