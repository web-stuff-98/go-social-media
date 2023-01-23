import classes from "../../../styles/FormClasses.module.scss";

const ErrorTip = ({ message }: { message: string }) => (
  <div className={classes.errorTip}>{message}</div>
);

export default ErrorTip;
