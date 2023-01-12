import classes from "../../../styles/components/chat/Rooms.module.scss";
import IconBtn from "../../IconBtn";
import { BiDoorOpen } from "react-icons/bi";
import { useState, useEffect } from "react";
import { getRoomImage } from "../../../services/rooms";
import useSocket from "../../../context/SocketContext";
import { useChat } from "./Chat";
import { IRoomCard } from "./Rooms";

export default function RoomCard({ r }: { r: IRoomCard }) {
  const { openSubscription, closeSubscription } = useSocket();
  const { openRoom } = useChat();
  const [imgURL, setImgURL] = useState("");

  useEffect(() => {
    if (r.img_blur) {
      setImgURL(r.img_blur);
      getRoomImage(r.ID)
        .then((url) => setImgURL(url))
        .catch(() => {});
    } else {
      setImgURL("");
    }
  }, [r.img_url]);

  useEffect(() => {
    openSubscription("room_card=" + r.ID);
    return () => {
      closeSubscription("room_card=" + r.ID);
    };
  }, []);

  return (
    <div
      style={
        r.img_blur
          ? {
              backgroundImage: `url(${imgURL || r.img_blur})`,
            }
          : {}
      }
      className={classes.room}
    >
      {r.name}
      <div className={classes.icons}>
        <IconBtn
          name="Enter room"
          ariaLabel="Enter room"
          style={{ color: "white" }}
          onClick={() => openRoom(r.ID)}
          Icon={BiDoorOpen}
        />
      </div>
    </div>
  );
}
