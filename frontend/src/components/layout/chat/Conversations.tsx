import classes from "../../../styles/components/chat/Conversations.module.scss";
import IconBtn from "../../IconBtn";
import { MdFileCopy, MdSend } from "react-icons/md";
import User from "../../User";
import { useUsers } from "../../../context/UsersContext";
import { useState, useEffect, useRef } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { IUser, useAuth } from "../../../context/AuthContext";
import PrivateMessage from "./PrivateMessage";
import useSocket from "../../../context/SocketContext";
import { instanceOfPrivateMessageData } from "../../../utils/DetermineSocketEvent";
import { useModal } from "../../../context/ModalContext";
import { getConversations, getConversation } from "../../../services/chat";
import ErrorTip from "../../ErrorTip";
import useFileSocket from "../../../context/FileSocketContext";

export interface IPrivateMessage {
  ID: string;
  uid: string;
  content: string;
  created_at: string;
  updated_at: string;
  has_attachment: boolean;
  recipient_id: string;
}
export type Conversation = {
  uid: string;
  messages: IPrivateMessage[];
};

export default function Conversations() {
  const { getUserData, cacheUserData } = useUsers();
  const { socket } = useSocket();
  const { uploadAttachment } = useFileSocket();
  const { user: currentUser } = useAuth();
  const { openModal } = useModal();

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState("");
  const selectedConversationRef = useRef("");
  const [selectedConversationIndex, setSelectedConversationIndex] =
    useState(-1);
  const selectedConversationIndexRef = useRef(-1);

  useEffect(() => {
    if (!currentUser) return;
    getConversations()
      .then((uids) => {
        if (uids) {
          uids.forEach((uid: string) => cacheUserData(uid));
          setConversations(uids.map((uid: string) => ({ uid, messages: [] })));
        }
      })
      .catch((e) => {
        openModal("Message", {
          msg: `${e}`,
          err: true,
          pen: false,
        });
      });
  }, [currentUser]);

  const openConversation = (uid: string) => {
    selectedConversationRef.current = uid;
    const i = conversations.findIndex((c) => c.uid === uid);
    selectedConversationIndexRef.current = i;
    setSelectedConversationIndex(i);
    setSelectedConversation(uid);
    getConversation(uid)
      .then((messages) => {
        setConversations((convs) => {
          let newConvs = convs;
          const i = convs.findIndex((c) => c.uid === uid);
          if (i === -1) {
            newConvs.push({ uid, messages });
          } else {
            newConvs[i] = { uid, messages };
          }
          return [...newConvs];
        });
      })
      .catch((e) => {
        openModal("Message", {
          msg: `${e}`,
          err: true,
          pen: false,
        });
      });
  };

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
        if (selectedConversationRef.current === data.DATA.recipient_id) {
          // If receiving own messages then put the
          // message inside the current conversation
          setConversations((conversations) => {
            let newConversations = conversations;
            newConversations[selectedConversationIndexRef.current].messages = [
              ...newConversations[selectedConversationIndexRef.current]
                .messages,
              data.DATA,
            ];
            return [...newConversations];
          });
        } else {
          // If there's no current conversation create one
          setConversations((conversations) => {
            let newConversations = conversations;
            newConversations.push({
              uid: data.DATA.recipient_id,
              messages: [data.DATA],
            });
            return [...newConversations];
          });
        }
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
    <>
      {user ? (
        <button
          key={user.ID}
          onClick={() => {
            openConversation(user.ID);
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
          <User small uid={user.ID} user={user} />
        </button>
      ) : (
        <></>
      )}
    </>
  );

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedConversation || messageInput.length > 200) return;
    socket?.send(
      JSON.stringify({
        event_type: "PRIVATE_MESSAGE",
        content: messageInput,
        recipient_id: selectedConversation,
      })
    );
    setMessageInput("");
  };

  const handleFile = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    window.alert("Uploading file")
    uploadAttachment(file, "000000000000000000000000");
  };

  const fileInputRef = useRef<HTMLInputElement>(null);
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
          <input ref={fileInputRef} type="file" onChange={handleFile} />
          <IconBtn
            name="Select attachment"
            ariaLabel="Select and attachment"
            Icon={MdFileCopy}
            onClick={() => fileInputRef.current?.click()}
          />
          <input
            value={messageInput}
            onChange={handleMessageInput}
            placeholder="Send a message..."
            type="text"
            required
          />
          <IconBtn name="Send" ariaLabel="Send message" Icon={MdSend} />
          {messageInput.length > 200 && (
            <ErrorTip message="Maximum 200 characters" />
          )}
        </form>
      </div>
    </>
  );
}
