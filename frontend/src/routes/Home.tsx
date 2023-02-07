import classes from "../styles/pages/SimplePage.module.scss";

export default function Home() {
  return (
    <div className={classes.container}>
      <h1>Welcome</h1>
      <hr />
      <p aria-live="assertive">
        New accounts are deleted automatically after 20 minutes. Logins use a
        simple username & password setup. You can create a new account or login
        using one of the example accounts. The usernames for the example
        accounts go through TestAcc1 up to TestAcc50, the password is Test1234!,
        make sure to include the capital T and the exclamation mark at the end.
      </p>
      <div className={classes.lists}>
        <div className={classes.listContainer}>
          <label>Website features</label>
          <ul>
            <li>Aria labelling & Tab indexing</li>
            <li>Private & group video chat</li>
            <li>Embedded comments</li>
            <li>Live updates for everything</li>
            <li>Invitations & Bans</li>
            <li>Customizable rooms</li>
            <li>File sharing with progress updates</li>
            <li>Lazy loading & placeholders</li>
            <li>Live blog with a rich text editor (quill)</li>
            <li>Pagination & sorting</li>
            <li>Filtering by terms and tags</li>
            <li>Voting on posts & comments</li>
          </ul>
        </div>
        <div className={classes.listContainer}>
          <label>Packages & Tech stuff</label>
          <ul>
            <li>RTL & Jest unit tests</li>
            <li>Intersection observers</li>
            <li>Image blur placeholders & lazy loading</li>
            <li>useCallback & useMemo</li>
            <li>startTransition & useDeferredValue</li>
            <li>Socket event models & interfaces</li>
            <li>Thread safe maps using mutex locks</li>
            <li>MongoDB with Changestreams</li>
            <li>Rate limiting middleware with redis</li>
            <li>React Context API & Reducers</li>
            <li>Refresh tokens with secure cookies</li>
            <li>Validation with Zod & Go validator v10</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
