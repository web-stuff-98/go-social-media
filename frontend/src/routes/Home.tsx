import classes from "../styles/pages/Home.module.scss";

export default function Home() {
  return (
    <div className={classes.container}>
      <h1>Welcome</h1>
      <hr />
      <p>
        This is my go social media project. You can create, update and edit blog
        posts, send direct messages, join and create chatrooms, comment on
        posts, comment on comments, vote on comments, vote on posts, search
        posts, sort posts, filter posts by tags and customize your profile
        image.
      </p>
      <div className={classes.listContainer}>
        <label>Packages and features</label>
        <ul>
          <li>Gorilla Mux & Gorilla Websocket</li>
          <li>Live updates through socket events</li>
          <li>MongoDB with Changestreams</li>
          <li>Rate limiting middleware with redis</li>
          <li>React Context API & Reducers</li>
          <li>Image blur placeholders & lazy loading</li>
          <li>Pagination, sorting & filtering</li>
          <li>Refresh tokens</li>
        </ul>
      </div>
    </div>
  );
}
