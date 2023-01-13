import classes from "./Layout.module.scss";
import { Link } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { useState } from "react";
import { useInterface } from "../../context/InterfaceContext";
import { MdMenu } from "react-icons/md";
import Dropdown from "../Dropdown";
import usePosts from "../../context/PostsContext";
import User from "../User";

export default function Nav() {
  const { user, logout } = useAuth();
  const {
    state: { isMobile },
  } = useInterface();
  const {
    getSortOrderFromParams,
    getSortModeFromParams,
    setSortModeInParams,
    setSortOrderInParams,
  } = usePosts();

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
            <span>Home</span>
          </Link>
          <Link to="/blog/1">
            <span>Blog</span>
          </Link>
          {!user && (
            <>
              <Link to="/login">
                <span>Login</span>
              </Link>
              <Link to="/register">
                <span>Register</span>
              </Link>
            </>
          )}
          <Link to="/policy">
            <span>Policy</span>
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
                <span>Logout</span>
              </button>
              <Link to="/editor">
                <span>Editor</span>
              </Link>
              <Link to="/settings">
                <span>Settings</span>
              </Link>
            </>
          )}
        </div>
      )}
      {user && <User reverse light small uid={user.ID} user={user}/>}
    </nav>
  );
}
