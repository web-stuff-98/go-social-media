import classes from "./Layout.module.scss";
import Nav from "./Nav";

import type { CSSProperties } from "react";

import { Outlet, useLocation } from "react-router-dom";
import { useInterface } from "../../context/InterfaceContext";
import { useMouse } from "../../context/MouseContext";
import Header from "./Header";
import Chat from "./chat/Chat";

export default function Layout() {
  const { state: iState } = useInterface();
  const { pathname } = useLocation();

  const mousePos = useMouse();

  const getStyle = () => {
    switch (iState.containerMode) {
      case "Feed": {
        const properties: CSSProperties = {
          width: "calc(100% - var(--horizontal-whitespace) * 2)",
          background: iState.isMobile ? "none" : "var(--foreground)",
          minHeight: "max-content",
          margin: "auto",
          display: "flex",
          ...(iState.isMobile
            ? {}
            : {
                borderLeft: "1px solid var(--base-pale)",
                borderRight: "1px solid var(--base-pale)",
                boxShadow: "0px 0px 3px rgba(0,0,0,0.166)",
              }),
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
          width: "20rem",
          height: "fit-content",
          maxWidth: "min(20rem, calc(100vw - 2rem))",
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
      <Chat />
      <main style={getStyle()}>
        <Outlet />
      </main>
    </div>
  );
}
