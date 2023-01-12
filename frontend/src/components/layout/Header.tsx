import { useAuth } from "../../context/AuthContext";
import { useInterface } from "../../context/InterfaceContext";
import User from "../User";

import classes from "./Layout.module.scss";

export default function Header() {
  const { state: iState, dispatch: iDispatch } = useInterface();
  const { user } = useAuth();

  return (
    <header>
      <div className={classes.logoImg}>
        <img
          alt="Go gopher - by Renee French. reneefrench.blogspot.com"
          src="go-mascot-wiki-colour.png"
        />
        Go-Social-Media
      </div>
      <div className={classes.right}>
        <button
          aria-label="Toggle dark mode"
          onClick={() => iDispatch({ darkMode: !iState.darkMode })}
          className={classes.darkToggle}
        >
          {iState.darkMode ? "Dark mode" : "Light mode"}
        </button>
        {user && (
          <div className={classes.userContainer}>
            <User light reverse uid={user.ID} user={user} />
          </div>
        )}
      </div>
    </header>
  );
}
