import { useEffect, useState, useCallback } from "react";
import { useModal } from "../../context/ModalContext";
import useSocket from "../../context/SocketContext";
import { useUsers } from "../../context/UsersContext";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
import {
  banFromRoom,
  getRoomPrivateData,
  unbanFromRoom,
} from "../../services/rooms";
import classes from "../../styles/components/chat/RoomMembers.module.scss";
import { instanceOfChangeData } from "../../utils/DetermineSocketEvent";
import ResMsg from "../shared/ResMsg";
import User from "../shared/User";
import { useChat } from "./Chat";

type RoomPrivateData = {
  members: string[];
  banned: string[];
};

export default function RoomMembers() {
  const { roomId } = useChat();
  const { cacheUserData, getUserData } = useUsers();
  const { openModal } = useModal();
  const { openSubscription, closeSubscription, socket } = useSocket();

  const [data, setData] = useState<RoomPrivateData>({
    members: [],
    banned: [],
  });
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const getData = async () => {
    try {
      setResMsg({ err: false, msg: "", pen: true });
      const data: RoomPrivateData = await getRoomPrivateData(roomId);
      data.members.forEach((uid) => cacheUserData(uid));
      data.banned.forEach((uid) => cacheUserData(uid));
      setData(data);
      setResMsg({ err: false, msg: "", pen: false });
    } catch (e) {
      setResMsg({ err: true, msg: `${e}`, pen: false });
    }
  };

  useEffect(() => {
    getData();
    openSubscription(`room_private_data=${roomId}`);
    return () => {
      closeSubscription(`room_private_data=${roomId}`);
    };
    // eslint-disable-next-line
  }, []);

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "MEMBER" || data.ENTITY === "BANNED") {
        const key = data.ENTITY === "MEMBER" ? "members" : "banned";
        const otherKey = data.ENTITY === "MEMBER" ? "banned" : "members";
        if (data.METHOD === "DELETE") {
          setData((o) => {
            let newData = o;
            newData[key].filter((uid) => uid !== data.DATA.ID);
            return newData;
          });
        }
        if (data.METHOD === "INSERT") {
          setData((o) => {
            let newData = o;
            newData[key].push(data.DATA.ID);
            newData[otherKey].filter((uid) => uid !== data.DATA.ID);
            return newData;
          });
        }
      }
    }
    // eslint-disable-next-line
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <div className={classes.container}>
      <label>Members ({data.members.length})</label>
      <ul aria-label="Members list">
        {data.members.map((uid) => (
          <li>
            <User square small uid={uid} user={getUserData(uid)} />
            <button
              onClick={async () => {
                try {
                  await banFromRoom(uid, roomId);
                } catch (e) {
                  openModal("Message", {
                    err: true,
                    msg: `${e}`,
                    pen: false,
                  });
                }
              }}
              aria-label="Ban user"
              className={classes.redButton}
            >
              Ban
            </button>
          </li>
        ))}
      </ul>
      <label>Banned ({data.banned.length})</label>
      <ul aria-label="Banned users list">
        {data.banned.map((uid) => (
          <li>
            <User square small uid={uid} user={getUserData(uid)} />
            <button
              onClick={async () => {
                try {
                  await unbanFromRoom(uid, roomId);
                } catch (e) {
                  openModal("Message", {
                    err: true,
                    msg: `${e}`,
                    pen: false,
                  });
                }
              }}
              aria-label="Unban user"
              className={classes.greenButton}
            >
              Unban
            </button>
          </li>
        ))}
      </ul>
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
