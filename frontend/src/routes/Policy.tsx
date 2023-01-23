import classes from "../styles/pages/SimplePage.module.scss";

export default function Policy() {
  return (
    <div className={classes.container}>
      <h1 style={{ marginBottom: "0" }}>Privacy & Cookies policy</h1>
      <hr style={{ marginTop: "0" }} />
      <p style={{ padding: "0" }}>
        This website uses cookies to facilitate logins. Accounts and all
        associated data are deleted automatically after 20 - 25 minutes. Your IP
        address is stored temporarily in order for the rate limiter to function,
        which guards against spam and repeated login attempts. None of your
        information is shared with any 3rd party or used for any reason other
        than as described here. It&apos;s a legal requirement to have a policy
        page on websites like this, although I could probably get away with not
        having one. You can watch the demo video instead if you don't want to
        login.
      </p>
    </div>
  );
}
