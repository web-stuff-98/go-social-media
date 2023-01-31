import { IPrivateMessage } from "../../interfaces/ChatInterfaces";
import classes from "../../styles/components/chat/PrivateMessage.module.scss";
import Attachment from "./Attachment";

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
  const DateTime = ({ dateString }: { dateString: string }) => (
    <>
      <span aria-label="Date" data-testid="Date">
        {dateString.split(", ")[0].replaceAll("/20", "/")}
      </span>
      <span aria-label="Time" data-testid="Time">
        {dateString.split(", ")[1]}
      </span>
    </>
  );

  return (
    <div
      data-testid="Container"
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      className={classes.container}
    >
      <div
        data-testid="Content container"
        style={reverse ? { textAlign: "right" } : {}}
        className={classes.content}
      >
        {msg.content}
        {msg.has_attachment && (
          <div className={classes.attachmentContainer}>
            <Attachment
              reverse={reverse}
              metaData={msg.attachment_metadata}
              progressData={
                msg.attachment_progress || {
                  failed: false,
                  pending: true,
                  ratio: 0,
                }
              }
              msgId={msg.ID}
            />
          </div>
        )}
      </div>
      <div
        data-testid="Date container"
        style={reverse ? { textAlign: "left" } : {}}
        className={classes.date}
      >
        <DateTime dateString={getDateString(new Date(msg.created_at))} />
      </div>
    </div>
  );
}
