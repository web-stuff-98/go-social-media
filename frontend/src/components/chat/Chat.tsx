import classes from "../../styles/components/chat/Chat.module.scss";
import Inbox from "./Inbox";
import { useState, useRef } from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import { useLocation } from "react-router-dom";
import Room from "./Room";
import { BsFillChatRightFill } from "react-icons/bs";
import ChatTopTray from "./ChatTopTray";
import RoomMembers from "./RoomMembers";
import useChat from "../../context/ChatContext";

export default function Chat() {
  const { pathname } = useLocation();
  const { notifications, section } = useChat();

  const [chatOpen, setChatOpen] = useState(false);

  const notificationsContainerRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  return (
    <div
      ref={containerRef}
      style={{
        ...(pathname.includes("/blog")
          ? {
              bottom:
                "calc(var(--pagination-controls) + var(--padding) + var(--padding))",
            }
          : {}),
        ...(chatOpen
          ? {}
          : { border: "none", background: "none", boxShadow: "none" }),
      }}
      className={classes.container}
    >
      {chatOpen ? (
        <>
          <ChatTopTray closeChat={() => setChatOpen(false)} />
          <div className={classes.inner}>
            {
              {
                Inbox: <Inbox />,
                Rooms: <Rooms />,
                Room: <Room />,
                Editor: <RoomEditor />,
                Menu: <Menu />,
                Members: <RoomMembers />,
              }[section]
            }
          </div>
        </>
      ) : (
        <button
          tabIndex={0}
          aria-hidden={chatOpen ? "true" : "false"}
          name="Open chat"
          aria-label="Open chat"
          onClick={() => setChatOpen(true)}
          type="button"
          className={classes.chatIconButton}
        >
          {notifications && notifications.length !== 0 && (
            <div
              ref={notificationsContainerRef}
              className={classes.notifications}
              style={{
                height: `${notificationsContainerRef.current?.clientWidth}px`,
                top: `-${
                  (notificationsContainerRef.current?.clientWidth || 0) * 0.33
                }px`,
                left: `-${
                  (notificationsContainerRef.current?.clientWidth || 0) * 0.33
                }px`,
              }}
            >
              +{notifications.length}
            </div>
          )}
          <BsFillChatRightFill />
        </button>
      )}
    </div>
  );
}
