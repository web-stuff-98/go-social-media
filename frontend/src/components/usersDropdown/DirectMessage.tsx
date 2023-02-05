import { UserMenuSection } from "../../context/UserdropdownContext";
import classes from "../../styles/components/shared/Userdropdown.module.scss";
import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import useSocket from "../../context/SocketContext";
import { MdSend } from "react-icons/md";

export default function DirectMessage({
  setDropdownSectionTo,
  closeUserDropdown,
  uid,
}: {
  setDropdownSectionTo: (section: UserMenuSection) => void;
  closeUserDropdown: () => void;
  uid: string;
}) {
  const { sendIfPossible } = useSocket();

  const privateMessageSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    sendIfPossible(
      JSON.stringify({
        event_type: "PRIVATE_MESSAGE",
        content: messageInput,
        recipient_id: uid,
      })
    );
    setMessageInput("");
    closeUserDropdown();
  };

  const [messageInput, setMessageInput] = useState("");

  return (
    <form className={classes.messageForm} onSubmit={privateMessageSubmit}>
      <input
        autoFocus
        value={messageInput}
        onChange={(e: ChangeEvent<HTMLInputElement>) =>
          setMessageInput(e.target.value)
        }
        placeholder="Direct message..."
        required
        type="text"
      />
      <button aria-label="Send direct message">
        <MdSend />
      </button>
    </form>
  );
}
