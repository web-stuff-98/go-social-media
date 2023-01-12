import { IRoomMessage } from "./Room";
import classes from "../../../styles/components/chat/Room.module.scss";
import User from "../../User";
import { useUsers } from "../../../context/UsersContext";

export default function RoomMessage({ msg }: { msg: IRoomMessage }) {
  const { getUserData } = useUsers();

  return (
    <div className={classes.message}>
      <User date={new Date(msg.created_at)} small uid={msg.uid} user={getUserData(msg.uid)} />
      <div className={classes.content}>{msg.content}</div>
    </div>
  );
}
