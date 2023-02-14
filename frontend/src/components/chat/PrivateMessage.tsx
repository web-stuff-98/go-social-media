import { IPrivateMessage } from "../../interfaces/ChatInterfaces";
import classes from "../../styles/components/chat/PrivateMessage.module.scss";
import Attachment from "./Attachment";
import { RiDeleteBin2Fill } from "react-icons/ri";
import { AiFillEdit } from "react-icons/ai";
import IconBtn from "../shared/IconBtn";
import { useAuth } from "../../context/AuthContext";
import { useState, useEffect, useRef, FormEvent } from "react";
import type { ChangeEvent } from "react";
import { useModal } from "../../context/ModalContext";
import useSocket from "../../context/SocketContext";
import { acceptInvite, declineInvite } from "../../services/rooms";

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
  const { user } = useAuth();
  const { sendIfPossible } = useSocket();
  const { openModal } = useModal();

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

  const [isEditing, setIsEditing] = useState(false);
  const [editInput, setEditInput] = useState("");
  const [mouseInside, setMouseInside] = useState(false);

  const handleSubmitUpdate = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editInput) {
      openModal("Message", {
        msg: "You cannot submit an empty message",
        err: true,
        pen: false,
      });
      setIsEditing(false);
    } else if (editInput.length > 200) {
      openModal("Message", {
        msg: "Max 200 characters",
        err: true,
        pen: false,
      });
    } else {
      sendIfPossible(
        JSON.stringify({
          event_type: "PRIVATE_MESSAGE_UPDATE",
          msg_id: msg.ID,
          content: editInput,
          recipient_id: msg.recipient_id,
        })
      );
      setIsEditing(false);
    }
  };

  const onClickOutside = () => {
    setIsEditing(false);
  };

  useEffect(() => {
    const clicked = () => {
      if (!mouseInside) {
        onClickOutside();
      }
    };
    document.addEventListener("mousedown", clicked);
    return () => {
      document.removeEventListener("mousedown", clicked);
    };
    // eslint-disable-next-line
  }, [mouseInside]);

  const textAreaRef = useRef<HTMLTextAreaElement>(null);
  return (
    <div
      data-testid="Container"
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      className={classes.container}
      onMouseEnter={() => setMouseInside(true)}
      onMouseLeave={() => setMouseInside(false)}
      tabIndex={0}
    >
      {msg.uid === user?.ID && !isEditing && (
        <div tabIndex={2} className={classes.icons}>
          {!msg.invitation && (
            <IconBtn
              type="button"
              name="Edit message"
              ariaLabel="Edit message"
              Icon={AiFillEdit}
              onClick={() => {
                setIsEditing(true);
                setMouseInside(true);
                setEditInput(msg.content);
              }}
            />
          )}
          <IconBtn
            type="button"
            name="Delete message"
            ariaLabel="Delete message"
            style={{ color: "red" }}
            Icon={RiDeleteBin2Fill}
            onClick={() =>
              openModal("Confirm", {
                msg: "Are you sure you want to delete this message?",
                err: false,
                pen: false,
                confirmationCallback: () =>
                  sendIfPossible(
                    JSON.stringify({
                      event_type: "PRIVATE_MESSAGE_DELETE",
                      msg_id: msg.ID,
                      recipient_id: msg.recipient_id,
                    })
                  ),
                cancellationCallback: () => {},
              })
            }
          />
        </div>
      )}
      <div
        data-testid="Content container"
        style={reverse ? { textAlign: "right" } : {}}
        className={classes.content}
        tabIndex={1}
      >
        {isEditing ? (
          <form onSubmit={handleSubmitUpdate}>
            <textarea
              autoFocus
              onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
                setEditInput(e.target.value)
              }
              value={editInput}
              ref={textAreaRef}
            />
            <button
              type="submit"
              aria-label="Update message"
              name="Update message"
            >
              Update
            </button>
          </form>
        ) : msg.invitation ? (
          msg.uid === user?.ID ? (
            !msg.invitation_accepted && !msg.invitation_declined ? (
              "Invitation sent"
            ) : (
              `Your invitation was ${
                msg.invitation_accepted ? "accepted" : "declined"
              }`
            )
          ) : !msg.invitation_accepted && !msg.invitation_declined ? (
            <>
              Invitation
              <br />
              <div className={classes.invitationButtons}>
                <button
                  //msg.content is the room id for invitations
                  onClick={() => acceptInvite(msg.uid, msg.ID, msg.content)}
                  aria-label="Accept invitation"
                  name="Accept invitation"
                >
                  ✅Accept
                </button>
                <button
                  //msg.content is the room id for invitations
                  onClick={() => declineInvite(msg.uid, msg.ID, msg.content)}
                  aria-label="Decline invitation"
                  name="Decline invitation"
                >
                  ❌Decline
                </button>
              </div>
            </>
          ) : (
            `Invitation ${msg.invitation_accepted ? "accepted" : "declined"}`
          )
        ) : (
          msg.content
        )}
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
