import classes from "../../styles/components/chat/Room.module.scss";
import User from "../shared/User";
import { useUsers } from "../../context/UsersContext";
import Attachment from "./Attachment";
import { IRoomMessage } from "../../interfaces/ChatInterfaces";
import IconBtn from "../shared/IconBtn";
import { RiDeleteBin2Fill } from "react-icons/ri";
import { AiFillEdit } from "react-icons/ai";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";
import useSocket from "../../context/SocketContext";
import { useEffect, useRef, useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import useChat from "../../context/ChatContext";

export default function RoomMessage({
  msg,
  reverse,
}: {
  msg: IRoomMessage;
  reverse: boolean;
}) {
  const { getUserData } = useUsers();
  const { user } = useAuth();
  const { openModal } = useModal();
  const { roomId } = useChat();
  const { sendIfPossible } = useSocket();

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
          event_type: "ROOM_MESSAGE_UPDATE",
          msg_id: msg.ID,
          content: editInput,
          room_id: roomId,
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
          ) : (
            msg.content
          )}
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
      {user?.ID === msg.uid && !isEditing && (
        <div className={classes.icons}>
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
                      event_type: "ROOM_MESSAGE_DELETE",
                      msg_id: msg.ID,
                      room_id: roomId,
                    })
                  ),
                cancellationCallback: () => {},
              })
            }
          />
        </div>
      )}
    </div>
  );
}
