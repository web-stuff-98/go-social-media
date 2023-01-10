import classes from "../../../styles/components/chat/Chat.module.scss";
import IconBtn from "../../IconBtn";
import { IoMdClose } from "react-icons/io";
import Conversations from "./Conversations";
import { MdMenu } from "react-icons/md";
import {
  useState,
  createContext,
  useContext,
  useEffect,
  useCallback,
} from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import useSocket from "../../../context/SocketContext";
import { instanceOfPrivateMessageData } from "../../../utils/DetermineSocketEvent";
import { useUsers } from "../../../context/UsersContext";

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
  inbox: Inbox;
}>({
  section: ChatSection.MENU,
  setSection: () => {},
  inbox: [],
});

export interface IPrivateMessage {
  ID: string;
  uid: string;
  content: string;
  created_at: string;
  updated_at: string;
  has_attachment: boolean;
  attachment_pending: boolean;
  attachment_type: string;
  attachment_error: boolean;
}
export type Conversation = {
  uid: string;
  messages: IPrivateMessage[];
};
export type Inbox = Conversation[];

export const useChat = () => useContext(ChatContext);

export default function Chat() {
  const { socket } = useSocket();
  const { cacheUserData } = useUsers();
  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);

  const [inbox, setInbox] = useState<Inbox>([]);

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    console.log("Cunt")
    if (instanceOfPrivateMessageData(data)) {
    console.log("Fuck")
      cacheUserData(data.DATA.uid);
      setInbox((inbox) => {
        let newInbox = inbox;
        const conversationIndex = inbox.findIndex(
          (c) => c.uid === data.DATA.uid
        );
        if (conversationIndex === -1) {
          newInbox.push({ uid: data.DATA.uid, messages: [data.DATA] });
        } else {
          newInbox[conversationIndex].messages = [
            ...newInbox[conversationIndex].messages,
            data.DATA,
          ];
        }
        return [...newInbox];
      });
    }
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <div className={classes.container}>
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
      <ChatContext.Provider value={{ section, setSection, inbox }}>
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
