export type CreateRoomFormData = {
  name: string;
  description: string;
};

export type CreateRoomFormErrors = Partial<
  Record<keyof CreateRoomFormData, string>
>;

export function validateCreateRoomForm(
  formData: CreateRoomFormData,
): CreateRoomFormErrors {
  const errors: CreateRoomFormErrors = {};
  const name = formData.name.trim();
  const description = formData.description.trim();

  if (name === "") {
    errors.name = "Room name is required.";
  } else if (name.length > 255) {
    errors.name = "Room name must be 255 characters or fewer.";
  }

  if (description.length > 500) {
    errors.description = "Description must be 500 characters or fewer.";
  }

  return errors;
}

export function hasCreateRoomFormErrors(errors: CreateRoomFormErrors): boolean {
  return Object.values(errors).some(Boolean);
}
