import classes from "../styles/pages/Home.module.scss";

export default function Home() {
  return (
    <div className={classes.container}>
      <h1>Welcome</h1>
      <hr />
      <p>
        This is my go social media project. Its my second Go project, and the
        new version of Go-Chat. You can make nested comments, edit posts, sort,
        paginate and filter posts, make and edit rooms, customize rooms,
        customize your profile and share files. Go-Chat is broken, but this
        project is complete and actually works.
      </p>
      <div className={classes.listContainer}>
        <label>Packages and features</label>
        <ul>
          <li>Gorilla Mux & Gorilla Websocket</li>
          <li>Socket event models and interfaces</li>
          <li>MongoDB with Changestreams</li>
          <li>Rate limiting middleware with redis</li>
          <li>React Context API & Reducers</li>
          <li>Image blur placeholders & lazy loading</li>
          <li>Pagination, sorting & filtering</li>
          <li>Refresh tokens with secure cookies</li>
          <li>Validation with Zod & Go validator v10</li>
          <li>Live updates through socket events</li>
        </ul>
      </div>
    </div>
  );
}
