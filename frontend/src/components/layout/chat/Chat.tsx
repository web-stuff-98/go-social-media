import classes from "../../../styles/components/chat/Chat.module.scss";
import IconBtn from "../../IconBtn";
import { IoMdClose } from "react-icons/io";
import Conversations from "./Conversations";
import { MdMenu } from "react-icons/md";
import { useState, createContext, useContext } from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import { useLocation } from "react-router-dom";
import Room from "./Room";

export enum ChatSection {
  "MENU" = "Menu",
  "CONVS" = "Conversations",
  "ROOMS" = "Rooms",
  "ROOM" = "Room",
  "EDITOR" = "Room editor",
}

const ChatContext = createContext<{
  section: ChatSection;
  setSection: (to: ChatSection) => void;
  openRoom: (id: string) => void;
  roomId: string;
}>({
  section: ChatSection.MENU,
  setSection: () => {},
  openRoom: () => {},
  roomId: "",
});

export const useChat = () => useContext(ChatContext);

export default function Chat() {
  const { pathname } = useLocation();
  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);
  const [roomId, setRoomId] = useState("");
  const openRoom = (id: string) => {
    setRoomId(id);
    setSection(ChatSection.ROOM);
  };

  return (
    <div
      style={
        pathname.includes("/blog")
          ? { bottom: "calc(var(--pagination-controls) + var(--padding))" }
          : {}
      }
      className={classes.container}
    >
      <div className={classes.topTray}>
        {section}
        <div className={classes.icons}>
          <IconBtn
            onClick={() => setSection(ChatSection.MENU)}
            name="Chat menu"
            ariaLabel="Chat menu"
            Icon={MdMenu}
          />
          <IconBtn name="Close chat" ariaLabel="Close chat" Icon={IoMdClose} />
        </div>
      </div>
      <ChatContext.Provider
        value={{
          section,
          setSection,
          roomId,
          openRoom,
        }}
      >
        <div className={classes.inner}>
          {
            {
              ["Conversations"]: <Conversations />,
              ["Rooms"]: <Rooms />,
              ["Room"]: <Room />,
              ["Room editor"]: <RoomEditor />,
              ["Menu"]: <Menu />,
            }[section]
          }
        </div>
      </ChatContext.Provider>
    </div>
  );
}
