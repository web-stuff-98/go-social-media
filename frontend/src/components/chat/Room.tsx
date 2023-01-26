import classes from "../../styles/components/chat/Room.module.scss";
import { useState, useEffect, useCallback, useRef } from "react";
import type { FormEvent, ChangeEvent } from "react";
import { getRoom } from "../../services/rooms";
import { ChatSection, useChat } from "./Chat";
import ResMsg, { IResMsg } from "../shared/ResMsg";
import { IRoomCard } from "./Rooms";
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
} from "../../utils/DetermineSocketEvent";
import { useAuth } from "../../context/AuthContext";
import { useUsers } from "../../context/UsersContext";
import ErrorTip from "../shared/forms/ErrorTip";
import useAttachment from "../../context/AttachmentContext";
import { useModal } from "../../context/ModalContext";
import { IAttachmentData, IMsgAttachmentProgress } from "./Attachment";
import VideoChat from "./VideoChat";

import * as process from "process";
(window as any).process = process;

export interface IRoomMessage {
  ID: string;
  uid: string;
  content: string;
  created_at: string;
  updated_at: string;
  has_attachment: boolean;
  attachment_progress?: IMsgAttachmentProgress;
  attachment_metadata?: IAttachmentData;
}

export interface IRoom extends IRoomCard {
  messages: IRoomMessage[];
}

export default function Room() {
  const { roomId, setSection } = useChat();
  const { socket, openSubscription, closeSubscription, sendIfPossible } = useSocket();
  const { user } = useAuth();
  const { cacheUserData } = useUsers();
  const { uploadAttachment } = useAttachment();
  const { openModal } = useModal();

  const fileRef = useRef<File>();
  const [file, setFile] = useState<File>();
  const [room, setRoom] = useState<IRoom>();
  const [vidChatOpen, setVidChatOpen] = useState(false)
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  useEffect(() => {
    openSubscription(`room=${roomId}`);
    setResMsg({ msg: "", err: false, pen: true });
    getRoom(roomId)
      .then((room: IRoom) => {
        setRoom({ ...room, messages: room.messages || [] });
        setResMsg({ msg: "", err: false, pen: false });
      })
      .catch((e) => setResMsg({ msg: `${e}`, err: true, pen: false }));
    return () => {
      closeSubscription(`room=${roomId}`);
    };
  }, []);

  const [messageInput, setMessageInput] = useState("");
  const handleMessageInput = (e: ChangeEvent<HTMLInputElement>) => {
    setMessageInput(e.target.value);
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (messageInput.length > 200) return;
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
        await uploadAttachment(fileRef.current, data.DATA.ID, roomId, true);
        fileRef.current = undefined;
      }
    }
    if (instanceOfChangeData(data)) {
      if (data.DATA.ID !== roomId) return;
      if (data.METHOD === "DELETE") {
        setSection(ChatSection.MENU);
      }
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
    }
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
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
  return (
    <div className={classes.container}>
      {room ? (
        <div className={classes.messagesAndVideoChat}>
          {vidChatOpen && <VideoChat isRoom id={room.ID} />}
          {room.messages.length > 0 ? (
            <div className={classes.messages}>
              {room.messages.map((msg) => (
                <RoomMessage
                  key={msg.ID}
                  reverse={msg.uid !== user?.ID}
                  msg={msg}
                />
              ))}
            </div>
          ) : (
            <p style={{textAlign:"center"}}>This room has recieved no messages.</p>
          )}
        </div>
      ) : (
        <></>
      )}
      <form onSubmit={handleSubmit} className={classes.messageForm}>
        <input ref={fileInputRef} type="file" onChange={handleFile} />
        <IconBtn
          name="Video chat"
          ariaLabel="Open video chat"
          type="button"
          onClick={() => setVidChatOpen(!vidChatOpen)}
          Icon={RiWebcamLine}
        />
        <IconBtn
          name="Send"
          ariaLabel="Send message"
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
          name="Send"
          ariaLabel="Send message"
          type="submit"
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
