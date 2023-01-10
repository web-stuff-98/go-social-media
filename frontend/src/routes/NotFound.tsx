import classes from "../styles/pages/NotFound.module.scss";

export default function NotFound() {
  return (
    <div className={classes.container}>
      <h1>404</h1>
      <hr />
      <p>Page not found</p>
    </div>
  );
}
