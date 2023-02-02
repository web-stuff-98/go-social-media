import classes from "../../styles/components/Layout.module.scss";
import { Link } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { useState } from "react";
import { useInterface } from "../../context/InterfaceContext";
import { MdMenu } from "react-icons/md";
import User from "../shared/User";
import { BsChevronCompactUp } from "react-icons/bs";

export default function Nav() {
  const { user, logout } = useAuth();
  const {
    state: { isMobile, darkMode },
  } = useInterface();

  const [mobileOpen, setMobileOpen] = useState(false);

  const NavLink = ({ to, name }: { to: string; name: string }) => (
    <Link to={to}>
      <span style={darkMode ? {} : { color: "black" }}>{name}</span>
    </Link>
  );
  const NavBtn = ({ name, onClick }: { name: string; onClick: Function }) => (
    <button
      onClick={() => onClick()}
      aria-label={name}
      name={name}
      style={{
        background: "none",
        border: "none",
        padding: "none",
      }}
    >
      <span style={darkMode ? {} : { color: "black" }}>Logout</span>
    </button>
  );

  return (
    <nav
      style={{
        ...(isMobile
          ? {
              justifyContent: "flex-end",
              alignItems: "center",
              ...(mobileOpen
                ? {
                    height: "16rem",
                    justifyContent: "space-between",
                    alignItems: "flex-end",
                    paddingLeft: "3px",
                  }
                : {}),
            }
          : {}),
      }}
    >
      {isMobile && (
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label={mobileOpen ? "Close nav links" : "Open nav links"}
          className={classes.menuIcon}
        >
          <MdMenu />
        </button>
      )}
      {(!isMobile || mobileOpen) && (
        <div
          style={
            mobileOpen && isMobile
              ? {
                  flexDirection: "column",
                  alignItems: "flex-start",
                  width: "fit-content",
                  marginBottom: "calc(var(--nav-height) + var(--padding))",
                }
              : {}
          }
          className={classes.navLinks}
        >
          <NavLink to="/" name="Home" />
          <NavLink to="/blog/1" name="Blog" />
          {!user && (
            <>
              <NavLink to="/login" name="Login" />
              <NavLink to="/register" name="Register" />
            </>
          )}
          <NavLink to="/policy" name="Policy" />
          {user && (
            <>
              <NavBtn name="Logout" onClick={logout} />
              <NavLink to="/editor" name="Editor" />
              <NavLink to="/settings" name="Settings" />
            </>
          )}
        </div>
      )}
      {user && (
        <User reverse light={darkMode} small uid={user.ID} user={user} />
      )}
    </nav>
  );
}
