import classes from "../../styles/components/Layout.module.scss";
import Nav from "./Nav";

import type { CSSProperties } from "react";

import { Outlet, useLocation } from "react-router-dom";
import { useInterface } from "../../context/InterfaceContext";
import { useMouse } from "../../context/MouseContext";
import Header from "./Header";
import Chat from "../chat/Chat";
import { useAuth } from "../../context/AuthContext";

export default function Layout() {
  const { state: iState } = useInterface();
  const { user } = useAuth();
  const { pathname } = useLocation();
  const mousePos = useMouse();

  const getStyle = () => {
    switch (iState.containerMode) {
      case "Feed": {
        const properties: CSSProperties = {
          width: "calc(100% - var(--horizontal-whitespace) * 2)",
          minHeight: pathname.includes("/blog") ? "max-content" : "100%",
          margin: "auto",
          display: "flex",
          position: "absolute",
          background: pathname.includes("/blog") ? "none" : "var(--foreground)",
          height: pathname.includes("/blog")
            ? "calc(100% - var(--header-height) - var(--nav-height) - var(--pagination-controls))"
            : "fit-content",
          left: "var(--horizontal-whitespace)",
          ...(!pathname.includes("/blog")
            ? {
                borderLeft: "1px solid var(--base-pale)",
                borderRight: "1px solid var(--base-pale)",
                boxShadow: "0px 0px 6px rgba(0,0,0,0.1)",
              }
            : {}),
        };
        return properties;
      }
      case "Full": {
        const properties: CSSProperties = {
          width: "100%",
        };
        return properties;
      }
      case "Modal": {
        const properties: CSSProperties = {
          width: "30rem",
          height: "fit-content",
          maxWidth: "min(30rem, calc(100vw - 2rem))",
          background: "var(--foreground)",
          border: "1px solid var(--base-medium)",
          borderRadius: "var(--border-radius)",
          margin: "auto",
          boxShadow: "2px 2px 3px rgba(0,0,0,0.066)",
        };
        return properties;
      }
    }
  };

  return (
    <div className={classes.container}>
      <div className={classes.backgroundOuterContainer}>
        <div className={classes.backgroundInnerContainer}>
          <div aria-label="hidden" className={classes.background} />
          <div
            aria-label="hidden"
            style={{
              maskImage: `radial-gradient(circle at ${
                (mousePos?.left! / iState.dimensions.width) * 100
              }% ${
                (mousePos?.top! / iState.dimensions.height) * 100
              }%, black 0%, transparent 7%)`,
              WebkitMaskImage: `radial-gradient(circle at ${
                (mousePos?.left! / iState.dimensions.width) * 100
              }% ${
                (mousePos?.top! / iState.dimensions.height) * 100
              }%, black 0%, transparent 7%)`,
              ...(iState.darkMode
                ? { filter: "brightness(4.7) contrast(1.8) blur(3px)" }
                : {}),
            }}
            className={classes.backgroundHover}
          />
        </div>
      </div>
      <Header />
      <Nav />
      {user && <Chat />}
      <main style={getStyle()}>
        <Outlet />
      </main>
    </div>
  );
}
