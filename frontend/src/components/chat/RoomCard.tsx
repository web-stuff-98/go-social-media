import classes from "../../styles/components/chat/Rooms.module.scss";
import IconBtn from "../shared/IconBtn";
import { BiDoorOpen } from "react-icons/bi";
import { AiFillEdit } from "react-icons/ai";
import { useState, useEffect } from "react";
import { getRoomImage } from "../../services/rooms";
import useSocket from "../../context/SocketContext";
import { useChat } from "./Chat";
import { useAuth } from "../../context/AuthContext";
import { IRoomCard } from "../../interfaces/ChatInterfaces";
import { RiDeleteBin2Fill } from "react-icons/ri";
import { IoPeople } from "react-icons/io5";

export default function RoomCard({ r }: { r: IRoomCard }) {
  const { user } = useAuth();
  const { openSubscription, closeSubscription } = useSocket();
  const { openRoom, openRoomEditor, deleteRoom, openRoomMembers } = useChat();
  const [imgURL, setImgURL] = useState("");

  const loadImage = async () => {
    try {
      setImgURL("");
      const url = await getRoomImage(r.ID);
      setImgURL(url);
    } catch (e) {
      console.log("Failed to load room image:", e);
    }
  };

  useEffect(() => {
    if (r.img_blur) {
      loadImage();
    } else {
      setImgURL("");
    }
    // eslint-disable-next-line
  }, [r.img_url]);

  useEffect(() => {
    openSubscription("room_card=" + r.ID);
    return () => {
      closeSubscription("room_card=" + r.ID);
    };
    // eslint-disable-next-line
  }, []);

  return (
    <div
      data-testid="Container"
      style={{
        ...(r.img_blur
          ? {
              backgroundImage: `url(${imgURL || r.img_blur})`,
            }
          : {}),
        ...(r.private && r.author_id !== user?.ID
          ? { filter: "opacity(0.5)", pointerEvents: "none" }
          : {}),
      }}
      className={classes.room}
    >
      {r.name}
      <div className={classes.icons}>
        {user && r.author_id === user.ID && (
          <>
            <IconBtn
              name="Room members"
              ariaLabel="Room members"
              onClick={() => openRoomMembers(r.ID)}
              style={{ color: "white" }}
              Icon={IoPeople}
            />
            <IconBtn
              name="Edit room"
              ariaLabel="Edit room"
              style={{ color: "white" }}
              onClick={() => openRoomEditor(r.ID)}
              Icon={AiFillEdit}
            />
            <IconBtn
              name="Delete room"
              ariaLabel="Delete room"
              style={{ color: "red" }}
              onClick={() => deleteRoom(r.ID)}
              Icon={RiDeleteBin2Fill}
            />
          </>
        )}
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
