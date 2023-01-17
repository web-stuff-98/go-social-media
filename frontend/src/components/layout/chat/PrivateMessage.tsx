import classes from "../../../styles/components/chat/PrivateMessage.module.scss";
import Attachment from "../../Attachment";
import { IPrivateMessage } from "./Conversations";

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "short",
  timeStyle: "short",
});

export default function PrivateMessage({
  msg,
  reverse,
}: {
  msg: IPrivateMessage;
  reverse: boolean;
}) {
  const getDateString = (date: Date) => dateFormatter.format(date);
  const renderDateTime = (dateString: string) => {
    return (
      <>
        <span>{dateString.split(", ")[0]}</span>
        <span>{dateString.split(", ")[1]}</span>
      </>
    );
  };

  return (
    <div
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      className={classes.container}
    >
      <div
        style={reverse ? { textAlign: "right" } : {}}
        className={classes.content}
      >
        {msg.content}
        {msg.has_attachment && (
          <div className={classes.attachmentContainer}>
            <Attachment
              progressData={
                msg.attachment_progress || { failed: false, pending: true, ratio: 0 }
              }
              msgId={msg.ID}
            />
          </div>
        )}
      </div>
      <div
        style={reverse ? { textAlign: "left" } : {}}
        className={classes.date}
      >
        {renderDateTime(getDateString(new Date(msg.created_at)))}
      </div>
    </div>
  );
}
