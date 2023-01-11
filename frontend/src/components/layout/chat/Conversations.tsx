import classes from "../../../styles/components/chat/Conversations.module.scss";
import IconBtn from "../../IconBtn";
import { MdSend } from "react-icons/md";
import User from "../../User";
import { useUsers } from "../../../context/UsersContext";
import { useState, useEffect, useRef } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { IUser, useAuth } from "../../../context/AuthContext";
import PrivateMessage from "./PrivateMessage";
import useSocket from "../../../context/SocketContext";
import { instanceOfPrivateMessageData } from "../../../utils/DetermineSocketEvent";
import { useModal } from "../../../context/ModalContext";
import { getConversations } from "../../../services/chat";

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

export default function Conversations() {
  const { getUserData, cacheUserData } = useUsers();
  const { socket } = useSocket();
  const { user: currentUser } = useAuth();
  const { openModal } = useModal();

  const [conversations, setConversations] = useState<Conversation[]>([]);

  useEffect(() => {
    if (!currentUser) return;
    getConversations()
      .then((data) => {
        const group: any = {};
        data.messages.forEach((msg: IPrivateMessage) => {
          if (group[msg.uid]) {
            group[msg.uid] = [...group[msg.uid], msg];
          } else {
            group[msg.uid] = [msg];
          }
        });
        Object.keys(group).forEach((k) => {
          setConversations((conversations) => [
            ...conversations,
            { uid: k, messages: group[k] },
          ]);
        });
      })
      .catch((e) => {
        openModal("Message", {
          msg: `${e}`,
          err: true,
          pen: false,
        });
      });
  }, [currentUser]);

  const [selectedConversation, setSelectedConversation] = useState("");
  const [selectedConversationIndex, setSelectedConversationIndex] =
    useState(-1);
  const selectedConversationIndexRef = useRef(-1);

  const handleMessage = (e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfPrivateMessageData(data)) {
      if (data.DATA.uid !== currentUser?.ID) {
        cacheUserData(data.DATA.uid);
        setConversations((conversations) => {
          let newConversations = conversations;
          const conversationIndex = conversations.findIndex(
            (c) => c.uid === data.DATA.uid
          );
          if (conversationIndex === -1) {
            newConversations.push({
              uid: data.DATA.uid,
              messages: [data.DATA],
            });
          } else {
            newConversations[conversationIndex].messages = [
              ...newConversations[conversationIndex].messages,
              data.DATA,
            ];
          }
          return [...newConversations];
        });
      } else {
        // If receiving own messages then put the
        // message inside the current conversation
        if (selectedConversationIndexRef.current === -1) return;
        setConversations((conversations) => {
          let newConversations = conversations;
          newConversations[selectedConversationIndexRef.current].messages = [
            ...newConversations[selectedConversationIndexRef.current].messages,
            data.DATA,
          ];
          return [...newConversations];
        });
      }
    }
  };

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  const [messageInput, setMessageInput] = useState("");
  const handleMessageInput = (e: ChangeEvent<HTMLInputElement>) =>
    setMessageInput(e.target.value);

  const renderConversee = (user: IUser) => (
    <button
      key={user.ID}
      onClick={() => {
        setSelectedConversation(user.ID);
        setSelectedConversationIndex(
          conversations.findIndex((c) => c.uid === user.ID)
        );
        selectedConversationIndexRef.current = conversations.findIndex(
          (c) => c.uid === user.ID
        );
      }}
      name={`Open conversation with ${user.ID}`}
      aria-label={`Open conversation with ${user.ID}`}
      style={
        selectedConversation === user.ID
          ? {
              background: "rgba(32,64,96,0.1666)",
            }
          : {}
      }
      className={classes.user}
    >
      <User uid={user.ID} user={user} />
    </button>
  );

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedConversation) return;
    socket?.send(
      JSON.stringify({
        event_type: "PRIVATE_MESSAGE",
        content: messageInput,
        recipient_id: selectedConversation,
      })
    );
    setMessageInput("");
  };

  return (
    <>
      <div className={classes.users}>
        {conversations.map((c) => renderConversee(getUserData(c.uid)))}
      </div>
      <div className={classes.right}>
        <div className={classes.messages}>
          {selectedConversation &&
            conversations[selectedConversationIndex].messages.map((msg) => (
              <PrivateMessage
                key={msg.ID}
                reverse={msg.uid !== currentUser?.ID}
                msg={msg}
              />
            ))}
        </div>
        <form onSubmit={handleSubmit} className={classes.messageForm}>
          <input
            value={messageInput}
            onChange={handleMessageInput}
            placeholder="Send a message..."
            type="text"
            required
          />
          <IconBtn name="Send" ariaLabel="Send message" Icon={MdSend} />
        </form>
      </div>
    </>
  );
}
