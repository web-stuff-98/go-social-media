import classes from "../../styles/components/Layout.module.scss";
import { Link } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { useEffect, useState } from "react";
import { useInterface } from "../../context/InterfaceContext";
import { MdMenu } from "react-icons/md";
import User from "../shared/User";

export default function Nav() {
  const { user, logout } = useAuth();
  const {
    state: { isMobile, darkMode },
  } = useInterface();

  const [mobileOpen, setMobileOpen] = useState(false);
  const [mobileLinksFadeIn, setMobileLinksFadeIn] = useState(false);

  const NavLink = ({ to, name }: { to: string; name: string }) => (
    <Link to={to}>
      <span style={darkMode ? {} : { color: "black" }}>{name}</span>
    </Link>
  );
  const NavBtn = ({ name, onClick }: { name: string; onClick: Function }) => (
    <button
      onClick={() => onClick()}
      aria-label={name === "Logout" ? "Log out" : name}
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

  useEffect(() => {
    if (mobileOpen) {
      setTimeout(() => {
        setMobileLinksFadeIn(mobileOpen);
      }, 200);
    } else {
      setMobileLinksFadeIn(false);
    }
  }, [mobileOpen]);

  return (
    <nav
      id="nav"
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
          aria-controls="Toggle navigation menu"
          onClick={() => setMobileOpen(!mobileOpen)}
          className={classes.menuIcon}
        >
          <MdMenu style={darkMode ? {} : { fill: "black" }} />
        </button>
      )}
      {(!isMobile || mobileOpen) && (
        <div
          aria-expanded={!isMobile || mobileOpen}
          style={{
            ...(mobileOpen && isMobile
              ? {
                  flexDirection: "column",
                  alignItems: "flex-start",
                  width: "fit-content",
                  marginBottom: "calc(var(--nav-height) + var(--padding))",
                }
              : {}),
            ...(isMobile
              ? {
                  filter: mobileLinksFadeIn ? "opacity(1)" : "opacity(0)",
                  transition: mobileLinksFadeIn ? "filter 200ms ease" : "none",
                }
              : {}),
          }}
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
        <User square reverse light={darkMode} small uid={user.ID} user={user} />
      )}
    </nav>
  );
}
