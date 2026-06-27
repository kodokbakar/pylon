import { describe, expect, it } from "vitest";

import {
  hasCreateRoomFormErrors,
  validateCreateRoomForm,
  type CreateRoomFormData,
} from "./room-validation";

const validForm: CreateRoomFormData = {
  name: "Backend Guild",
  description: "Room for backend discussions.",
};

describe("room validation", () => {
  it("accepts valid create room data", () => {
    const errors = validateCreateRoomForm(validForm);

    expect(errors).toEqual({});
    expect(hasCreateRoomFormErrors(errors)).toBe(false);
  });

  it("requires room name", () => {
    const errors = validateCreateRoomForm({
      ...validForm,
      name: "   ",
    });

    expect(errors.name).toBe("Room name is required.");
    expect(hasCreateRoomFormErrors(errors)).toBe(true);
  });

  it("limits description length", () => {
    const errors = validateCreateRoomForm({
      ...validForm,
      description: "a".repeat(501),
    });

    expect(errors.description).toBe(
      "Description must be 500 characters or fewer.",
    );
    expect(hasCreateRoomFormErrors(errors)).toBe(true);
  });
});
