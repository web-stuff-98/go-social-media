import classes from "../styles/pages/NotFound.module.scss";

export default function NotFound() {
  return (
    <div className={classes.container}>
      <h1 data-testid="heading">404</h1>
      <hr />
      <p data-testid="paragraph">Page not found</p>
    </div>
  );
}
