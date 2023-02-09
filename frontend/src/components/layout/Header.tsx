import { useInterface } from "../../context/InterfaceContext";
import classes from "../../styles/components/Layout.module.scss";

import Gopher from "./go-mascot-wiki-colour.png";

export default function Header() {
  const { state: iState, dispatch: iDispatch } = useInterface();

  return (
    <header>
      <div className={classes.logoImg}>
        <img
          alt="Go gopher - by Renee French. reneefrench.blogspot.com"
          src={Gopher}
        />
        <div className={classes.text}>
          <h1>Go-Social-Media</h1>
          <h2>By Jason</h2>
        </div>
      </div>
      <div className={classes.right}>
        <button
          aria-label="Toggle dark mode"
          aria-controls="Toggle dark mode"
          onClick={() => iDispatch({ darkMode: !iState.darkMode })}
          className={classes.darkToggle}
        >
          {iState.darkMode ? "Dark mode" : "Light mode"}
        </button>
      </div>
    </header>
  );
}
