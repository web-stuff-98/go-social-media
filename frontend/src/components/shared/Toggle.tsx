import classes from "../../styles/components/shared/Toggle.module.scss";

export default function Toggle({
  label = "Toggle",
  toggledOn,
  setToggledOn = () => {},
}: {
  label?: string;
  toggledOn: boolean;
  setToggledOn: (to: boolean) => void;
}) {
  return (
    <div className={classes.container}>
      <label htmlFor={label}>{label}</label>
      <button
        type="button"
        onClick={() => setToggledOn(!toggledOn)}
        aria-label={label}
        id={label}
        className={classes.container}
      >
        <span className={classes.sliderBar} />
        <span
          style={toggledOn ? { left: "calc(100% - 1rem)" } : {}}
          className={classes.toggler}
        />
        <span aria-hidden="true" className={classes.offLabel}>
          Off
        </span>
        <span aria-hidden="true" className={classes.onLabel}>
          On
        </span>
      </button>
    </div>
  );
}
