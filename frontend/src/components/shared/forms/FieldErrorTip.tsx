import ErrorTip from "./ErrorTip";
import { Fragment } from "react";

// ValidationErr is based on what formik outputs. path[0] should be the ID of the field.
export type ValidationErr = {
  path: string[];
  message: string;
};

export default function FieldErrorTip({
  validationErrs,
  fieldName,
}: {
  validationErrs: any[];
  fieldName: string;
}) {
  return validationErrs.findIndex((e: any) => e.path[0] === fieldName) !==
    -1 ? (
    <ErrorTip
      message={
        validationErrs.find((e: any) => e.path[0] === fieldName)
          ?.message as string
      }
    />
  ) : (
    <Fragment />
  );
}
