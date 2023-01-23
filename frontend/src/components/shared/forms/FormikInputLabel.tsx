import classes from "../../../styles/FormClasses.module.scss";
import { capitalize } from "lodash";
import type { ChangeEventHandler, FocusEventHandler } from "react";
import FieldErrorTip, { ValidationErr } from "./FieldErrorTip";

/**
 * Saves some space and makes code look cleaner
 */

const FormikInputAndLabel = ({
  name,
  id,
  ariaLabel,
  value,
  touched,
  validationErrs,
  onChange,
  onBlur,
}: {
  name: string;
  id: string;
  ariaLabel: string;
  value: string;
  touched?: boolean;
  validationErrs: ValidationErr[];
  onChange: ChangeEventHandler<HTMLInputElement>;
  onBlur: FocusEventHandler<HTMLInputElement>;
}) => (
  <div className={classes.inputLabelWrapper}>
    <label htmlFor={id}>{capitalize(name)}</label>
    <input
      data-testid={id}
      name={name}
      id={id}
      aria-label={ariaLabel}
      value={value}
      onChange={onChange}
      onBlur={onBlur}
    />
    {touched && (
      <FieldErrorTip fieldName={id} validationErrs={validationErrs} />
    )}
  </div>
);

export default FormikInputAndLabel;
