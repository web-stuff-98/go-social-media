import { useRef, useEffect } from "react";
import type { ReactElement } from "react";
import { useAuth } from "../../context/AuthContext";
import { AiOutlineUser } from "react-icons/ai";
import { useUsers } from "../../context/UsersContext";
import classes from "../../styles/components/shared/User.module.scss";
import { useUserdropdown } from "../../context/UserdropdownContext";
import { IUser } from "../../interfaces/GeneralInterfaces";

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "short",
  timeStyle: "short",
});

export default function User({
  user = { username: "Username", ID: "123" },
  uid = "",
  onClick = undefined,
  date,
  reverse,
  light,
  AdditionalStuff,
  small,
  testid,
  square,
}: {
  user?: IUser;
  uid: string;
  onClick?: () => void;
  date?: Date;
  reverse?: boolean;
  light?: boolean;
  AdditionalStuff?: ReactElement;
  small?: boolean;
  testid?: string;
  square?: boolean;
}) {
  const { userEnteredView, cacheUserData, userLeftView } = useUsers();
  const { openUserdropdown } = useUserdropdown();
  const { user: currentUser } = useAuth();
  const containerRef = useRef<HTMLDivElement>(null);

  const observer = new IntersectionObserver(([entry]) => {
    // why uid === "undefined", don't remember, could probably be removed
    if (!uid || uid === "undefined") return;
    if (entry.isIntersecting) {
      userEnteredView(uid);
      cacheUserData(uid);
    } else {
      userLeftView(uid);
    }
  });
  useEffect(() => {
    observer.observe(containerRef.current!);
    return () => {
      observer.disconnect();
    };
    // eslint-disable-next-line
  }, []);

  const getDateString = (date: Date) => dateFormatter.format(date);
  const renderDateTime = (dateString: string) => {
    return (
      <div
        style={reverse ? { alignItems: "flex-end", textAlign: "right" } : {}}
        className={classes.dateTime}
      >
        <span style={light ? { color: "white" } : {}}>
          {dateString.split(", ")[0]}
        </span>
        <span style={light ? { color: "white" } : {}}>
          {dateString.split(", ")[1]}
        </span>
      </div>
    );
  };

  return (
    <div
      data-testid={testid}
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      ref={containerRef}
      className={classes.container}
    >
      {user && (
        <>
          <button
            tabIndex={0}
            role="button"
            aria-label={`Open menu on user ${user.username}`}
            style={{
              ...(user.base64pfp
                ? { backgroundImage: `url(${user.base64pfp})` }
                : {}),
              ...(onClick || (currentUser && currentUser.ID !== user.ID)
                ? { cursor: "pointer" }
                : {}),
              ...(light ? { border: "1px solid white" } : {}),
              ...(small
                ? {
                    width: "1.866rem",
                    height: "1.866rem",
                    minWidth: "1.866rem",
                    minHeight: "1.866rem",
                  }
                : {}),
              ...(square
                ? {
                    borderRadius: "var(--border-radius-medium)",
                  }
                : {}),
            }}
            onClick={() => {
              if (onClick) onClick();
              else if (currentUser && currentUser.ID !== user.ID)
                openUserdropdown(user.ID);
            }}
            className={classes.pfp}
          >
            {!user.base64pfp && (
              <AiOutlineUser
                style={light ? { fill: "white" } : {}}
                className={classes.pfpIcon}
              />
            )}
          </button>
          {AdditionalStuff && (
            <div className={classes.additionalStuff}>{AdditionalStuff}</div>
          )}
          <div
            style={{
              ...(light ? { color: "white" } : {}),
              ...(reverse
                ? { textAlign: "right", alignItems: "flex-end" }
                : {}),
            }}
            className={classes.text}
          >
            <div
              style={{
                ...(light ? { color: "white" } : {}),
                ...(small
                  ? {
                      fontSize: "0.833rem",
                      lineHeight: "0.9",
                    }
                  : {}),
              }}
              className={classes.name}
            >
              {user.username}
            </div>
            {date && renderDateTime(getDateString(date))}
          </div>
        </>
      )}
      {!user && AdditionalStuff && (
        <div className={classes.additionalStuff}>{AdditionalStuff}</div>
      )}
    </div>
  );
}
