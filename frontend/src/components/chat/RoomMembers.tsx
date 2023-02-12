import { useEffect, useState, useCallback } from "react";
import useChat from "../../context/ChatContext";
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
    const controller = new AbortController();
    getData();
    openSubscription(`room_private_data=${roomId}`);
    return () => {
      closeSubscription(`room_private_data=${roomId}`);
      controller.abort();
    };
    // eslint-disable-next-line
  }, []);

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "MEMBER") {
        if (data.METHOD === "DELETE") {
          setData((o) => ({
            members: [...o.members.filter((uid) => uid !== data.DATA.ID)],
            banned: o.banned,
          }));
        }
        if (data.METHOD === "INSERT") {
          setData((o) => ({
            members: [...o.members, data.DATA.ID],
            banned: [...o.banned.filter((uid) => uid !== data.DATA.ID)],
          }));
        }
      }
      if (data.ENTITY === "BANNED") {
        if (data.METHOD === "DELETE") {
          setData((o) => ({
            banned: [...o.banned.filter((uid) => uid !== data.DATA.ID)],
            members: o.members,
          }));
        }
        if (data.METHOD === "INSERT") {
          setData((o) => ({
            banned: [...o.banned, data.DATA.ID],
            members: [...o.members.filter((uid) => uid !== data.DATA.ID)],
          }));
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
    // eslint-disable-next-line
  }, [socket]);

  return (
    <div className={classes.container}>
      <label htmlFor="Members list">Members ({data.members.length})</label>
      <ul aria-label="Members list" id="Members list">
        {data.members.map((uid) => (
          <li>
            <User square small uid={uid} user={getUserData(uid)} />
            <button
              onClick={() => {
                openModal("Confirm", {
                  msg: "Are you sure you want to ban this user?",
                  err: false,
                  pen: false,
                  confirmationCallback: async () => {
                    try {
                      await banFromRoom(uid, roomId);
                      openModal("Message", {
                        err: false,
                        msg: "User has been banned",
                        pen: false,
                      });
                    } catch (e) {
                      openModal("Message", {
                        err: true,
                        msg: `${e}`,
                        pen: false,
                      });
                    }
                  },
                });
              }}
              aria-label="Ban user"
              className={classes.redButton}
            >
              Ban
            </button>
          </li>
        ))}
      </ul>
      <label htmlFor="Banned list">Banned ({data.banned.length})</label>
      <ul aria-label="Banned list" id="Banned list">
        {data.banned.map((uid) => (
          <li>
            <User square small uid={uid} user={getUserData(uid)} />
            <button
              onClick={() => {
                openModal("Confirm", {
                  msg: "Are you sure you want to unban this user?",
                  err: false,
                  pen: false,
                  confirmationCallback: async () => {
                    try {
                      await unbanFromRoom(uid, roomId);
                      openModal("Message", {
                        err: false,
                        msg: "User has been unbanned, if your room is private you will need to invite them again for them to join",
                        pen: false,
                      });
                    } catch (e) {
                      openModal("Message", {
                        err: true,
                        msg: `${e}`,
                        pen: false,
                      });
                    }
                  },
                });
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
