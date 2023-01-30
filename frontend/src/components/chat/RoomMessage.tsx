import { IRoomMessage } from "./Room";
import classes from "../../styles/components/chat/Room.module.scss";
import User from "../shared/User";
import { useUsers } from "../../context/UsersContext";
import Attachment from "./Attachment";

export default function RoomMessage({
  msg,
  reverse,
}: {
  msg: IRoomMessage;
  reverse: boolean;
}) {
  const { getUserData } = useUsers();

  return (
    <div
      data-testid="Message container"
      style={
        reverse
          ? {
              flexDirection: "row-reverse",
            }
          : {}
      }
      className={classes.message}
    >
      <User
        reverse={reverse}
        date={new Date(msg.created_at)}
        small
        uid={msg.uid}
        user={getUserData(msg.uid)}
      />
      <div className={classes.content}>
        <div
          data-testid="Text container"
          style={reverse ? { textAlign: "right" } : {}}
          className={classes.text}
        >
          {msg.content}
        </div>
        {msg.has_attachment && (
          <Attachment
            reverse={reverse}
            msgId={msg.ID}
            progressData={msg.attachment_progress!}
            metaData={msg.attachment_metadata}
          />
        )}
      </div>
    </div>
  );
}
