import classes from "../../styles/components/Layout.module.scss";
import { Link } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { useState } from "react";
import { useInterface } from "../../context/InterfaceContext";
import { MdMenu } from "react-icons/md";
import User from "../shared/User";

export default function Nav() {
  const { user, logout } = useAuth();
  const {
    state: { isMobile, darkMode },
  } = useInterface();

  const [mobileOpen, setMobileOpen] = useState(false);

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
          <Link to="/">
            <span style={darkMode ? {} : { color: "black" }}>Home</span>
          </Link>
          <Link to="/blog/1">
            <span style={darkMode ? {} : { color: "black" }}>Blog</span>
          </Link>
          {!user && (
            <>
              <Link to="/login">
                <span style={darkMode ? {} : { color: "black" }}>Login</span>
              </Link>
              <Link to="/register">
                <span style={darkMode ? {} : { color: "black" }}>Register</span>
              </Link>
            </>
          )}
          <Link to="/policy">
            <span style={darkMode ? {} : { color: "black" }}>Policy</span>
          </Link>
          {user && (
            <>
              <button
                onClick={() => logout()}
                aria-label="Logout"
                style={{
                  background: "none",
                  border: "none",
                  padding: "none",
                }}
              >
                <span style={darkMode ? {} : { color: "black" }}>Logout</span>
              </button>
              <Link to="/editor">
                <span style={darkMode ? {} : { color: "black" }}>Editor</span>
              </Link>
              <Link to="/settings">
                <span style={darkMode ? {} : { color: "black" }}>Settings</span>
              </Link>
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
