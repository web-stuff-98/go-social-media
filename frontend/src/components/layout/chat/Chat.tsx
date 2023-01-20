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
  "EDITOR" = "Editor",
}

const ChatContext = createContext<{
  section: ChatSection;
  setSection: (to: ChatSection) => void;
  openRoom: (id: string) => void;
  openRoomEditor: (id: string) => void;
  roomId: string;
  editRoomId: string;
}>({
  section: ChatSection.MENU,
  setSection: () => {},
  openRoom: () => {},
  openRoomEditor: () => {},
  roomId: "",
  editRoomId: "",
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

  const [editRoomId, setEditRoomId] = useState("");
  const openRoomEditor = (id: string) => {
    setEditRoomId(id);
    setSection(ChatSection.EDITOR);
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
          editRoomId,
          openRoom,
          openRoomEditor,
        }}
      >
        <div className={classes.inner}>
          {
            {
              Conversations: <Conversations />,
              Rooms: <Rooms />,
              Room: <Room />,
              Editor: <RoomEditor />,
              Menu: <Menu />,
            }[section]
          }
        </div>
      </ChatContext.Provider>
    </div>
  );
}
