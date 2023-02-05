import classes from "../../styles/components/shared/Userdropdown.module.scss";
import { useEffect, useState } from "react";
import { IRoomCard } from "../../interfaces/ChatInterfaces";
import { getOwnRooms, banFromRoom } from "../../services/rooms";
import { useModal } from "../../context/ModalContext";

export default function BanFromRoom({
  closeUserDropdown,
  uid,
}: {
  closeUserDropdown: () => void;
  uid: string;
}) {
  const { openModal } = useModal();

  const [rooms, setRooms] = useState<IRoomCard[]>([]);

  const getRooms = async () => {
    try {
      const rooms = await getOwnRooms();
      setRooms(rooms);
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
      closeUserDropdown();
    }
  };

  const ban = async (inviteTo: string) => {
    try {
      await banFromRoom(uid, inviteTo);
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
  }, []);

  return (
    <div className={classes.inviteToRoom}>
      <ul id="rooms list" className={classes.roomsList}>
        <label htmlFor="rooms list">Ban user from room</label>
        {rooms.map((room) => (
          <li>
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
    </div>
  );
}
