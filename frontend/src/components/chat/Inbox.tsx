import classes from "../../styles/components/chat/Inbox.module.scss";
import IconBtn from "../shared/IconBtn";
import { MdFileCopy, MdSend } from "react-icons/md";
import User from "../shared/User";
import { useUsers } from "../../context/UsersContext";
import { useState, useEffect, useRef, useCallback } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { useAuth } from "../../context/AuthContext";
import PrivateMessage from "./PrivateMessage";
import useSocket from "../../context/SocketContext";
import {
  instanceOfAttachmentCompleteData,
  instanceOfAttachmentProgressData,
  instanceOfPrivateMessageData,
} from "../../utils/DetermineSocketEvent";
import { useModal } from "../../context/ModalContext";
import { getConversations, getConversation } from "../../services/chat";
import ErrorTip from "../shared/forms/ErrorTip";
import useAttachment from "../../context/AttachmentContext";
import VideoChat from "./VideoChat";
import { RiWebcamLine } from "react-icons/ri";
import { IConversation } from "../../interfaces/ChatInterfaces";
import { useChat } from "./Chat";

export default function Inbox() {
  const { getUserData, cacheUserData } = useUsers();
  const { socket, sendIfPossible } = useSocket();
  const { uploadAttachment } = useAttachment();
  const { user: currentUser } = useAuth();
  const { openModal } = useModal();
  const { notifications, toggleStream } = useChat();

  const [inbox, setInbox] = useState<IConversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState("");
  const [selectedConversationIndex, setSelectedConversationIndex] =
    useState(-1);
  const selectedConversationIndexRef = useRef(-1);
  const selectedConversationRef = useRef("");

  const loadConvs = async () => {
    try {
      const uids = await getConversations();
      uids.forEach((uid: string) => cacheUserData(uid));
      setInbox(uids.map((uid: string) => ({ uid, messages: [] })));
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
    }
  };

  useEffect(() => {
    if (!currentUser) return;
    loadConvs();
    // eslint-disable-next-line
  }, [currentUser]);

  useEffect(() => {
    return () => {
      if (selectedConversationRef.current)
        sendIfPossible(
          JSON.stringify({
            event_type: "EXIT_CONV",
            uid: selectedConversationRef.current,
          })
        );
    };
    // eslint-disable-next-line
  }, []);

  const openConversation = async (uid: string) => {
    try {
      const messages = await getConversation(uid);
      setInbox((convs) => {
        let newConvs = convs;
        const i = convs.findIndex((c) => c.uid === uid);
        if (i === -1) {
          newConvs.push({ uid, messages });
        } else {
          newConvs[i] = { uid, messages };
        }
        return [...newConvs];
      });
      if (
        selectedConversationRef.current &&
        selectedConversationRef.current !== uid
      ) {
        sendIfPossible(
          JSON.stringify({
            event_type: "EXIT_CONV",
            uid: selectedConversationRef.current,
          })
        );
      }
      selectedConversationRef.current = uid;
      const i = inbox.findIndex((c) => c.uid === uid);
      selectedConversationIndexRef.current = i;
      setSelectedConversationIndex(i);
      setSelectedConversation(uid);
      sendIfPossible(
        JSON.stringify({
          event_type: "OPEN_CONV",
          uid,
        })
      );
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
    }
  };

  const handleMessage = useCallback(async (e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfPrivateMessageData(data)) {
      if (data.DATA.uid !== currentUser?.ID) {
        cacheUserData(data.DATA.uid);
        setInbox((inbox) => {
          let newInbox = inbox;
          const conversationIndex = inbox.findIndex(
            (c) => c.uid === data.DATA.uid
          );
          if (conversationIndex === -1) {
            newInbox.push({
              uid: data.DATA.uid,
              messages: [data.DATA],
            });
          } else {
            newInbox[conversationIndex].messages = [
              ...newInbox[conversationIndex].messages,
              data.DATA,
            ];
          }
          return [...newInbox];
        });
      } else {
        // If recieving own message and an attachment was selected, upload it
        if (selectedConversationRef.current === data.DATA.recipient_id) {
          // If receiving own messages then put the
          // message inside the current conversation
          setInbox((inbox) => {
            let newInbox = inbox;
            newInbox[selectedConversationIndexRef.current].messages = [
              ...newInbox[selectedConversationIndexRef.current].messages,
              data.DATA,
            ];
            return [...newInbox];
          });
        } else {
          // If there's no current conversation create one
          setInbox((inbox) => {
            let newInbox = inbox;
            newInbox.push({
              uid: data.DATA.recipient_id,
              messages: [data.DATA],
            });
            return [...newInbox];
          });
        }
        if (fileRef.current) {
          setFile(undefined);
          uploadAttachment(
            fileRef.current,
            data.DATA.ID,
            data.DATA.recipient_id,
            false
          );
          fileRef.current = undefined;
        }
      }
    }
    if (instanceOfAttachmentProgressData(data)) {
      if (selectedConversationIndexRef.current !== -1) {
        setInbox((inbox) => {
          let newInbox = inbox;
          const i = inbox[
            selectedConversationIndexRef.current
          ].messages.findIndex((msg) => msg.ID === data.DATA.ID);
          if (i !== -1) {
            newInbox[selectedConversationIndexRef.current].messages[
              i
            ].attachment_progress = {
              failed: data.DATA.failed,
              pending: data.DATA.pending,
              ratio: data.DATA.ratio,
            };
            return [...newInbox];
          } else {
            return inbox;
          }
        });
      }
    }
    if (instanceOfAttachmentCompleteData(data)) {
      if (selectedConversationIndexRef.current !== -1) {
        setInbox((inbox) => {
          let newInbox = inbox;
          const i = inbox[
            selectedConversationIndexRef.current
          ].messages.findIndex((msg) => msg.ID === data.DATA.ID);
          if (i !== -1) {
            newInbox[selectedConversationIndexRef.current].messages[
              i
            ].attachment_progress = {
              failed: false,
              pending: false,
              ratio: 1,
            };
            newInbox[selectedConversationIndexRef.current].messages[
              i
            ].attachment_metadata = {
              size: data.DATA.size,
              type: data.DATA.type,
              name: data.DATA.name,
              length: data.DATA.length,
            };
            return [...newInbox];
          } else {
            return inbox;
          }
        });
      }
    }
    // eslint-disable-next-line
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  const [messageInput, setMessageInput] = useState("");
  const handleMessageInput = (e: ChangeEvent<HTMLInputElement>) =>
    setMessageInput(e.target.value);

  const handleSubmit = (e?: FormEvent<HTMLFormElement>) => {
    e?.preventDefault();
    if (!selectedConversation || messageInput.length > 200 || !messageInput)
      return;
    sendIfPossible(
      JSON.stringify({
        event_type: "PRIVATE_MESSAGE",
        content: messageInput,
        recipient_id: selectedConversation,
        has_attachment: file ? true : false,
      })
    );
    setMessageInput("");
  };

  const [file, setFile] = useState<File>();
  const fileRef = useRef<File>();
  const handleFile = async (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    if (file.size > 1024 * 1024 * 20) {
      openModal("Message", {
        msg: "File too large, Max 20mb",
        err: true,
        pen: false,
      });
      return;
    }
    setFile(file);
    fileRef.current = file;
  };

  const fileInputRef = useRef<HTMLInputElement>(null);
  return (
    <>
      <div data-testid="Users list" className={classes.users}>
        {inbox.map((c) => (
          <button
            key={c.uid}
            onClick={() => {
              openConversation(c.uid as string);
            }}
            name={`Open conversation with ${c.uid}`}
            aria-label={`Open conversation with ${c.uid}`}
            style={
              selectedConversation === c.uid
                ? {
                    background: "rgba(32,64,96,0.1666)",
                  }
                : {}
            }
            className={classes.user}
          >
            <User
              testid={`conversation uid:${c.uid}`}
              small
              uid={c.uid}
              user={getUserData(c.uid)}
            />
            {notifications &&
              notifications.map((n) => n.type.includes(c.uid)).length > 0 && (
                <div className={classes.notifications}>
                  +{notifications.map((n) => n.type.includes(c.uid)).length}
                </div>
              )}
          </button>
        ))}
      </div>
      <div className={classes.right}>
        <div
          data-testid="Messages and video chat"
          className={classes.messagesAndVideoChat}
        >
          {selectedConversation && <VideoChat id={selectedConversation} />}
          <div className={classes.messages}>
            {selectedConversation &&
              inbox[selectedConversationIndex].messages.map((msg) => (
                <PrivateMessage
                  key={msg.ID}
                  reverse={msg.uid !== currentUser?.ID}
                  msg={msg}
                />
              ))}
          </div>
        </div>
        <form
          data-testid="Message form"
          onSubmit={handleSubmit}
          className={classes.messageForm}
        >
          <input ref={fileInputRef} type="file" onChange={handleFile} />
          <IconBtn
            name="Video chat"
            ariaLabel="Open video chat"
            type="button"
            onClick={toggleStream}
            Icon={RiWebcamLine}
          />
          <IconBtn
            name="Select attachment"
            ariaLabel="Select an attachment"
            Icon={MdFileCopy}
            style={
              file
                ? { color: "lime", filter: "drop-shadow(0px,2px,1px,black)" }
                : {}
            }
            onClick={() => fileInputRef.current?.click()}
          />
          <input
            data-testid="Message input"
            name="Message input"
            value={messageInput}
            onChange={handleMessageInput}
            placeholder="Send a message..."
            type="text"
            required
          />
          <IconBtn
            onClick={() => handleSubmit()}
            type="button"
            testid="Send button"
            name="Send"
            ariaLabel="Send message"
            Icon={MdSend}
          />
          {messageInput.length > 200 && (
            <ErrorTip message="Maximum 200 characters" />
          )}
        </form>
      </div>
    </>
  );
}
