import classes from "../../styles/components/chat/Room.module.scss";
import { useState, useEffect, useCallback, useRef } from "react";
import type { FormEvent, ChangeEvent } from "react";
import { getRoom } from "../../services/rooms";
import { ChatSection, useChat } from "./Chat";
import ResMsg from "../shared/ResMsg";
import RoomMessage from "./RoomMessage";
import useSocket from "../../context/SocketContext";
import { MdFileCopy, MdSend } from "react-icons/md";
import { RiWebcamLine } from "react-icons/ri";
import IconBtn from "../shared/IconBtn";
import {
  instanceOfAttachmentCompleteData,
  instanceOfAttachmentProgressData,
  instanceOfChangeData,
  instanceOfRoomMessageData,
  instanceOfRoomMessageDeleteData,
  instanceOfRoomMessageUpdateData,
} from "../../utils/DetermineSocketEvent";
import { useAuth } from "../../context/AuthContext";
import { useUsers } from "../../context/UsersContext";
import ErrorTip from "../shared/forms/ErrorTip";
import useAttachment from "../../context/AttachmentContext";
import { useModal } from "../../context/ModalContext";
import VideoChat from "./VideoChat";
import { IoPeople } from "react-icons/io5";

import * as process from "process";
import { IRoom } from "../../interfaces/ChatInterfaces";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
(window as any).process = process;

export default function Room() {
  const { roomId, setSection, toggleStream, openRoomMembers } = useChat();
  const { socket, openSubscription, closeSubscription, sendIfPossible } =
    useSocket();
  const { user } = useAuth();
  const { cacheUserData } = useUsers();
  const { uploadAttachment } = useAttachment();
  const { openModal } = useModal();

  const fileRef = useRef<File>();
  const [file, setFile] = useState<File>();
  const [room, setRoom] = useState<IRoom>();
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const loadRoom = async () => {
    setResMsg({ msg: "", err: false, pen: true });
    try {
      const room = await getRoom(roomId);
      setRoom({ ...room, messages: room.messages || [] });
      setResMsg({ msg: "", err: false, pen: false });
    } catch (e) {
      setResMsg({ msg: `${e}`, err: true, pen: false });
    }
  };

  useEffect(() => {
    openSubscription(`room=${roomId}`);
    const controller = new AbortController();
    loadRoom();
    return () => {
      closeSubscription(`room=${roomId}`);
      controller.abort();
    };
    // eslint-disable-next-line
  }, []);

  const [messageInput, setMessageInput] = useState("");
  const handleMessageInput = (e: ChangeEvent<HTMLInputElement>) => {
    setMessageInput(e.target.value);
  };

  const handleSubmit = (e?: FormEvent<HTMLFormElement>) => {
    e?.preventDefault();
    if (messageInput.length > 200 || !messageInput) return;
    sendIfPossible(
      JSON.stringify({
        event_type: "ROOM_MESSAGE",
        content: messageInput,
        room_id: roomId,
        has_attachment: file ? true : false,
      })
    );
    setMessageInput("");
  };

  useEffect(() => {
    messagesBottomRef.current?.scrollIntoView({ behavior: "auto" });
  }, [room?.messages]);

  const handleMessage = useCallback(async (e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfRoomMessageData(data)) {
      if (data.DATA.uid !== user?.ID) cacheUserData(data.DATA.uid);
      setRoom((o) => {
        if (!o) return o;
        return { ...o, messages: [...o.messages, data.DATA] };
      });
      if (data.DATA.uid === user?.ID && fileRef.current) {
        setFile(undefined);
        uploadAttachment(fileRef.current, data.DATA.ID, roomId, true);
        fileRef.current = undefined;
      }
      return;
    }
    if (instanceOfChangeData(data)) {
      if (data.DATA.ID !== roomId) return;
      if (data.METHOD === "DELETE") {
        setSection(ChatSection.MENU);
      }
      return;
    }
    if (instanceOfRoomMessageDeleteData(data)) {
      setRoom((o) => {
        if (!o) return o;
        return {
          ...o,
          messages: [...o.messages.filter((msg) => msg.ID !== data.DATA.ID)],
        };
      });
      return;
    }
    if (instanceOfRoomMessageUpdateData(data)) {
      setRoom((o) => {
        if (!o) return o;
        let newMsgs = o.messages;
        const i = o.messages.findIndex((msg) => msg.ID === data.DATA.ID);
        if (i === -1) return o;
        newMsgs[i].content = data.DATA.content;
        return {
          ...o,
          messages: newMsgs,
        };
      });
      return;
    }
    if (instanceOfAttachmentProgressData(data)) {
      setRoom((o) => {
        if (!o) return o;
        const i = o.messages.findIndex((m) => m.ID === data.DATA.ID);
        if (i !== -1) {
          let newRoom = o;
          newRoom.messages[i].attachment_progress = {
            pending: data.DATA.pending,
            failed: data.DATA.failed,
            ratio: data.DATA.ratio,
          };
          return { ...newRoom };
        } else {
          return o;
        }
      });
      return;
    }
    if (instanceOfAttachmentCompleteData(data)) {
      setRoom((o) => {
        if (!o) return o;
        const i = o.messages.findIndex((m) => m.ID === data.DATA.ID);
        if (i !== -1) {
          let newRoom = o;
          newRoom.messages[i].attachment_progress = {
            pending: false,
            failed: false,
            ratio: 1,
          };
          newRoom.messages[i].attachment_metadata = {
            size: data.DATA.size,
            type: data.DATA.type,
            name: data.DATA.name,
            length: data.DATA.length,
          };
          return { ...newRoom };
        } else {
          return o;
        }
      });
      return;
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
  const messagesBottomRef = useRef<HTMLDivElement>(null);
  return (
    <div className={classes.container}>
      {room ? (
        <div
          data-testid="Messages and videochat"
          className={classes.messagesAndVideoChat}
        >
          <VideoChat isRoom id={roomId} />
          {room.messages.length > 0 ? (
            <div className={classes.messages}>
              {room.messages.map((msg) => (
                <RoomMessage
                  key={msg.ID}
                  reverse={msg.uid !== user?.ID}
                  msg={msg}
                />
              ))}
              <div className={classes.messagesBottomRef} />
            </div>
          ) : (
            <p style={{ textAlign: "center" }}>
              The room has recieved no messages.
            </p>
          )}
        </div>
      ) : (
        <></>
      )}
      <form
        data-testid="Message form"
        onSubmit={handleSubmit}
        className={classes.messageForm}
      >
        <input ref={fileInputRef} type="file" onChange={handleFile} />
        <IconBtn
          name="View users"
          ariaLabel="View users"
          type="button"
          onClick={() => {
            openRoomMembers(roomId);
          }}
          Icon={IoPeople}
        />
        <IconBtn
          name="Video chat"
          ariaLabel="Open video chat"
          type="button"
          onClick={() => {
            toggleStream(true, roomId);
          }}
          Icon={RiWebcamLine}
        />
        <IconBtn
          name="Select file"
          ariaLabel="Select file"
          type="button"
          Icon={MdFileCopy}
          style={
            file
              ? { color: "lime", filter: "drop-shadow(0px,2px,1px,black)" }
              : {}
          }
          onClick={() => fileInputRef.current?.click()}
        />
        <input
          value={messageInput}
          onChange={handleMessageInput}
          placeholder="Send a message..."
          type="text"
          required
        />
        <IconBtn
          testid="Send message"
          onClick={() => handleSubmit()}
          name="Send button"
          ariaLabel="Send message"
          type="button"
          Icon={MdSend}
        />
        {messageInput.length > 200 && (
          <ErrorTip message="Maximum 200 characters" />
        )}
      </form>
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
