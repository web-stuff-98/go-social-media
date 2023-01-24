import { useState } from "react";
import { ValidationErr } from "../components/shared/forms/FieldErrorTip";

export default function useFormikValidate(schema: any) {
  const [validationErrs, setValidationErrs] = useState<ValidationErr[]>([]);

  const validate = (vals: object) => {
    if (!schema) return;
    try {
      schema.parse(vals);
      setValidationErrs([]);
    } catch (e: any) {
      setValidationErrs(e.issues);
    }
  };

  return { validate, validationErrs };
}
