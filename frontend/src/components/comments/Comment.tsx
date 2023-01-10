import { useUsers } from "../../context/UsersContext";
import classes from "../../styles/components/Comments.module.scss";
import IconBtn from "../IconBtn";
import User from "../User";
import { AiFillEdit } from "react-icons/ai";
import { MdDelete } from "react-icons/md";
import {
  deleteComment,
  submitComment,
  updateComment,
} from "../../services/posts";
import { useState } from "react";
import { FaReply } from "react-icons/fa";
import ErrorTip from "../ErrorTip";
import { CommentForm } from "./CommentForm";

export interface IComment {
  ID: string;
  content: string;
  author_id: string;
  created_at: string;
  updated_at: string;
  parent_id: string;
}

export default function Comment({
  comment,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
}: {
  comment: IComment;
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
  getReplies: (parentId: string) => IComment[];
}) {
  const { getUserData } = useUsers();
  const [err, setErr] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [repliesOpen, setRepliesOpen] = useState(false);
  const childComments = getReplies(comment.ID);

  return (
    <div className={classes.container}>
      <div className={classes.top}>
        <User
          iconBtns={[
            <IconBtn
              type="button"
              ariaLabel="Edit comment"
              name="edit"
              onClick={() => setIsEditing(true)}
              Icon={AiFillEdit}
            />,
            <IconBtn
              type="button"
              ariaLabel="Delete comment"
              name="delete"
              style={{ color: "red" }}
              onClick={() =>
                deleteComment(postId, comment.ID).catch((e) => setErr(`${e}`))
              }
              Icon={MdDelete}
            />,
          ]}
          date={comment.created_at ? new Date(comment.created_at) : new Date()}
          uid={comment.author_id}
          user={getUserData(comment.author_id)}
        />
        <div className={classes.content}>
          {isEditing ? (
            <CommentForm
              autoFocus
              initialValue={comment.content}
              onSubmit={(c: string) => {
                updateComment(postId, comment.ID, c).catch((e) => setErr(e));
              }}
              onClickOutside={() => setIsEditing(false)}
              placeholder="Edit comment..."
            />
          ) : (
            <>{comment.content}</>
          )}
          {err && <ErrorTip message={err} />}
        </div>
        <div className={classes.icons}>
          <IconBtn
            onClick={() => setReplyingTo(comment.ID)}
            Icon={FaReply}
            name="Reply to comment..."
            ariaLabel="Reply to comment..."
          />
        </div>
      </div>
      {replyingTo === comment.ID && (
        <div className={classes.commentForm}>
          <CommentForm
            autoFocus
            onSubmit={(c: string) => submitComment(c, postId, replyingTo)}
            onClickOutside={() => setReplyingTo("")}
            placeholder="Reply to comment..."
          />
        </div>
      )}
      {childComments && repliesOpen && (
        <button
          onClick={() => setRepliesOpen(false)}
          className={classes.repliesBar}
        >
          <span />
        </button>
      )}
      {childComments && repliesOpen && childComments.length > 0 && (
        <div className={classes.childComments}>
          {childComments.map((cmt: IComment) => (
            <Comment
              getReplies={getReplies}
              setReplyingTo={setReplyingTo}
              replyingTo={replyingTo}
              postId={postId}
              comment={cmt}
            />
          ))}
        </div>
      )}
      {!repliesOpen && childComments && childComments.length > 0 && (
        <button
          onClick={() => setRepliesOpen(true)}
          className={classes.showRepliesButton}
        >
          Show replies
        </button>
      )}
    </div>
  );
}
