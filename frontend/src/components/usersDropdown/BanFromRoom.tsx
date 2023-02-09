import classes from "../../styles/components/shared/Userdropdown.module.scss";
import { useEffect, useState } from "react";
import { IRoomCard } from "../../interfaces/ChatInterfaces";
import { getOwnRooms, banFromRoom } from "../../services/rooms";
import { useModal } from "../../context/ModalContext";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
import ResMsg from "../shared/ResMsg";

export default function BanFromRoom({
  closeUserDropdown,
  uid,
  adjust,
}: {
  closeUserDropdown: () => void;
  uid: string;
  adjust: (delay: boolean) => void;
}) {
  const { openModal } = useModal();

  const [rooms, setRooms] = useState<IRoomCard[]>([]);
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const getRooms = async () => {
    try {
      setResMsg({ msg: "", err: false, pen: true });
      const rooms = await getOwnRooms();
      setRooms(rooms);
      if (rooms.length === 0) {
        setResMsg({ msg: "You have no rooms", err: true, pen: false });
      } else {
        setResMsg({ msg: "", err: false, pen: false });
      }
      adjust(true);
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
      setResMsg({ msg: "", err: false, pen: false });
      closeUserDropdown();
    }
  };

  const ban = async (inviteTo: string) => {
    try {
      await banFromRoom(uid, inviteTo);
      openModal("Message", {
        msg: "The user is now banned from the room",
        err: false,
        pen: false,
      });
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
    } finally {
      closeUserDropdown();
    }
  };

  useEffect(() => {
    getRooms();
    // eslint-disable-next-line
  }, []);

  return (
    <div className={classes.inviteToRoom}>
      <ul id="rooms list" className={classes.roomsList}>
        <label htmlFor="rooms list">Ban user from room</label>
        {rooms.map((room) => (
          <li key={room.ID}>
            <button
              onClick={() => ban(room.ID)}
              style={
                room.img_blur
                  ? { backgroundImage: `url(${room.img_blur})` }
                  : {}
              }
              className={classes.room}
            >
              {room.name}
            </button>
          </li>
        ))}
      </ul>
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
