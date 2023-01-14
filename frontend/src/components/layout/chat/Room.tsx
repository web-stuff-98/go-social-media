import classes from "../../../styles/components/chat/Room.module.scss";
import { useState, useEffect, useCallback } from "react";
import type { FormEvent, ChangeEvent } from "react";
import { getRoom } from "../../../services/rooms";
import { ChatSection, useChat } from "./Chat";
import ResMsg, { IResMsg } from "../../ResMsg";
import { IRoomCard } from "./Rooms";
import RoomMessage from "./RoomMessage";
import useSocket from "../../../context/SocketContext";
import { MdSend } from "react-icons/md";
import IconBtn from "../../IconBtn";
import {
  instanceOfChangeData,
  instanceOfRoomMessageData,
} from "../../../utils/DetermineSocketEvent";
import { useAuth } from "../../../context/AuthContext";
import { useUsers } from "../../../context/UsersContext";

export interface IRoomMessage {
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

export interface IRoom extends IRoomCard {
  messages: IRoomMessage[];
}

export default function Room() {
  const { roomId, setSection } = useChat();
  const { socket, openSubscription, closeSubscription } = useSocket();
  const { user } = useAuth();
  const { cacheUserData } = useUsers();

  const [room, setRoom] = useState<IRoom>();
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
    socket?.send(
      JSON.stringify({
        event_type: "ROOM_MESSAGE",
        content: messageInput,
        room_id: roomId,
      })
    );
    setMessageInput("");
  };

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfRoomMessageData(data)) {
      if (data.DATA.uid !== user?.ID) cacheUserData(data.DATA.uid);
      setRoom((o) => {
        if (!o) return o;
        return { ...o, messages: [...o.messages, data.DATA] };
      });
      return
    }
    if (instanceOfChangeData(data)) {
      if (data.DATA.ID !== roomId) return;
      if (data.METHOD === "DELETE") {
        setSection(ChatSection.MENU);
      }
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
      {room ? (
        <div className={classes.messages}>
          {room.messages.length > 0 ? (
            <>
              {room.messages.map((msg) => (
                <RoomMessage msg={msg} />
              ))}
            </>
          ) : (
            <p>This room has recieved no messages.</p>
          )}
        </div>
      ) : (
        <></>
      )}
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
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
