import classes from "../styles/pages/Home.module.scss";

export default function Policy() {
  return (
    <div className={classes.container}>
      <h1 style={{ marginBottom: "0" }}>Privacy & Cookies policy</h1>
      <hr style={{ marginTop: "0" }} />
      <p style={{ padding: "0" }}>
        This website uses cookies to facilitate user logins. Accounts and all
        associated data are deleted automatically after 20 minutes. Your IP
        address is stored locally in memory in order for the rate limiter to
        function, which guards against spam and repeated login attempts. No
        information is stored with any 3rd party or used for any reason other
        than as described here.
      </p>
    </div>
  );
}
