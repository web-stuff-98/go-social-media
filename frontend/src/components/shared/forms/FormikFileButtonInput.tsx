import classes from "../../../styles/FormClasses.module.scss";
import { capitalize } from "lodash";
import { useRef } from "react";
import type { ChangeEvent } from "react";
import FieldErrorTip, { ValidationErr } from "./FieldErrorTip";
import { FormikErrors } from "formik";

/**
 * Saves some space and makes code look cleaner...
 * This one is for file inputs
 */

const FormikFileButtonInput = ({
  name,
  id,
  ariaLabel,
  touched,
  accept,
  validationErrs,
  setFieldValue,
  setURL,
  setOriginalChanged,
  showLabel,
}: {
  name: string;
  id: string;
  ariaLabel: string;
  touched?: boolean;
  accept: string;
  validationErrs: ValidationErr[];
  setFieldValue: (
    field: string,
    value: any,
    shouldValidate?: boolean
  ) => Promise<void> | Promise<FormikErrors<object>>;
  setURL?: (to: string) => void;
  setOriginalChanged?: (to: boolean) => void;
  showLabel?: boolean;
}) => {
  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    setFieldValue(id, file);
    if (setURL) setURL(URL.createObjectURL(file));
    if (setOriginalChanged) setOriginalChanged(true);
  };

  const inputRef = useRef<HTMLInputElement>(null);
  return (
    <div className={classes.inputLabelWrapper}>
      {showLabel && <label htmlFor={id}>{capitalize(name)}</label>}
      <input
        onChange={handleChange}
        name={name}
        id={id}
        type="file"
        aria-label={ariaLabel}
        accept={accept}
        ref={inputRef}
      />
      <button onClick={() => inputRef.current?.click()} type="button">
        Select {name}
      </button>
      {touched && (
        <FieldErrorTip fieldName={id} validationErrs={validationErrs} />
      )}
    </div>
  );
};
export default FormikFileButtonInput;
