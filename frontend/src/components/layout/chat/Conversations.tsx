import classes from "../../../styles/components/chat/Conversations.module.scss";
import IconBtn from "../../IconBtn";
import { MdSend } from "react-icons/md";
import { useChat } from "./Chat";
import User from "../../User";
import { useUsers } from "../../../context/UsersContext";

export default function Conversations() {
  const { getUserData } = useUsers();
  const { inbox } = useChat();

  return (
    <>
      <div className={classes.users}>
        {inbox.map((c) => (
          <User uid={c.uid} user={getUserData(c.uid)} />
        ))}
      </div>
      <div className={classes.right}>
        <div className={classes.messages}></div>
        <form className={classes.messageForm}>
          <input placeholder="Send a message..." type="text" required />
          <IconBtn name="Send" ariaLabel="Send message" Icon={MdSend} />
        </form>
      </div>
    </>
  );
}
