import { useRef, useLayoutEffect, ReactElement } from "react";
import IconBtn from "./IconBtn";
import { IUser } from "../context/AuthContext";
import { AiOutlineUser } from "react-icons/ai";
import { useUsers } from "../context/UsersContext";
import classes from "../styles/components/User.module.scss";

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
  iconBtns,
}: {
  user?: IUser;
  uid: string;
  onClick?: () => void;
  date?: Date;
  reverse?: boolean;
  light?: boolean;
  iconBtns?: ReactElement[];
}) {
  const { userEnteredView, cacheUserData, userLeftView } = useUsers();
  const containerRef = useRef<HTMLDivElement>(null);

  const observer = new IntersectionObserver(([entry]) => {
    if (!uid || uid === "undefined") return;
    if (entry.isIntersecting) {
      userEnteredView(uid);
      cacheUserData(uid);
    } else {
      userLeftView(uid);
    }
  });
  useLayoutEffect(() => {
    observer.observe(containerRef.current!);
    return () => {
      if (uid) userLeftView(uid);
      observer.disconnect();
    };
  }, [containerRef.current]);

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
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      ref={containerRef}
      className={classes.container}
    >
      {user && (
        <>
          <span
            style={{
              ...(user.base64pfp
                ? { backgroundImage: `url(${user.base64pfp})` }
                : {}),
              ...(onClick ? { cursor: "pointer" } : {}),
              ...(light ? { border: "1px solid white" } : {}),
            }}
            onClick={() => {
              if (onClick) onClick();
            }}
            className={classes.pfp}
          >
            {!user.base64pfp && (
              <AiOutlineUser
                style={light ? { fill: "white" } : {}}
                className={classes.pfpIcon}
              />
            )}
          </span>
          {iconBtns && (
            <div className={classes.iconBtns}>{iconBtns.map((btn) => btn)}</div>
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
              style={light ? { color: "white" } : {}}
              className={classes.name}
            >
              {user.username}
            </div>
            {date && renderDateTime(getDateString(date))}
          </div>
        </>
      )}
    </div>
  );
}
