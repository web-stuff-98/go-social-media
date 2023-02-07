import classes from "../../styles/components/shared/ProgressBar.module.scss";

/**
 * Ratio = 0 - 1
 */

export default function ProgressBar({ ratio }: { ratio: number }) {
  return (
    <div
      aria-hidden="true"
      data-testid="Progress bar"
      aria-label={`${Math.floor(ratio * 100)}% completed`}
      className={classes.container}
    >
      <span aria-hidden="true">
        <span aria-hidden="true" style={{ width: `${ratio * 100}%` }} />
      </span>
    </div>
  );
}
