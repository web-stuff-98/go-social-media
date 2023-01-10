import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { MdSend } from "react-icons/md";
import { ImSpinner8 } from "react-icons/im";
import ErrorTip from "../ErrorTip";
import classes from "../../styles/components/CommentForm.module.scss";

export function CommentForm({
  loading,
  error,
  onSubmit,
  autoFocus = false,
  initialValue = "",
  placeholder = "",
  onClickOutside = () => {},
}: {
  loading?: boolean;
  error?: string;
  onSubmit: Function;
  autoFocus?: boolean;
  initialValue?: string;
  placeholder?: string;
  onClickOutside?: Function;
}) {
  const [message, setMessage] = useState(initialValue);

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    onSubmit(message);
    setMessage("")
  }

  const [mouseInside, setMouseInside] = useState(false);

  const onMouseEnter = () => setMouseInside(true);
  const onMouseLeave = () => setMouseInside(false);

  useEffect(() => {
    const clicked = () => {
      if (!mouseInside) {
        onClickOutside();
      }
    };
    document.addEventListener("mousedown", clicked);
    return () => {
      document.removeEventListener("mousedown", clicked);
    };
  }, [mouseInside]);

  return (
    <form className={classes.container} onSubmit={handleSubmit}>
      <div className={classes.inner}>
        <input
          autoFocus={autoFocus}
          value={message}
          placeholder={placeholder}
          onChange={(e) => setMessage(e.target.value)}
          onMouseEnter={() => onMouseEnter()}
          onMouseLeave={() => onMouseLeave()}
        />
        <button type="submit" aria-label="Submit" disabled={loading}>
          {loading ? (
            <ImSpinner8 className={classes.spinner} />
          ) : (
            <MdSend className={classes.sendIcon} />
          )}
        </button>
        {error && <ErrorTip message={String(error)} />}
      </div>
    </form>
  );
}
