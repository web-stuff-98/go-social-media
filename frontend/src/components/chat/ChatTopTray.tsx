import classes from "../../styles/components/chat/Chat.module.scss";
import IconBtn from "../shared/IconBtn";
import { MdMenu } from "react-icons/md";
import { IoMdClose } from "react-icons/io";
import useChat, { ChatSection } from "../../context/ChatContext";

export default function ChatTopTray({ closeChat }: { closeChat: () => void }) {
  const { section, setSection } = useChat();
  return (
    <div className={classes.topTray}>
      {section === "Editor" ? "Room editor" : section}
      <div className={classes.icons}>
        {section !== ChatSection.MENU && (
          <IconBtn
            onClick={() => setSection(ChatSection.MENU)}
            name="Chat menu"
            ariaLabel="Go back to chat menu"
            Icon={MdMenu}
          />
        )}
        <IconBtn
          name="Close chat"
          ariaLabel="Close chat"
          onClick={() => closeChat()}
          Icon={IoMdClose}
        />
      </div>
    </div>
  );
}
