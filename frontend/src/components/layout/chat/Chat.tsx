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
}>({
  section: ChatSection.MENU,
  setSection: () => {},
});

export const useChat = () => useContext(ChatContext);

export default function Chat() {
  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);
  const { pathname } = useLocation();

  return (
    <div style={pathname.includes("/blog") ? {bottom:"calc(var(--pagination-controls) + var(--padding))"} : {}} className={classes.container}>
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
        }}
      >
        <div className={classes.inner}>
          {
            {
              ["Conversations"]: <Conversations />,
              ["Rooms"]: <Rooms />,
              ["Room"]: <Conversations />,
              ["Room editor"]: <RoomEditor />,
              ["Menu"]: <Menu />,
            }[section]
          }
        </div>
      </ChatContext.Provider>
    </div>
  );
}
