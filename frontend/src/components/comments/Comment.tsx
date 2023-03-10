import { useUsers } from "../../context/UsersContext";
import classes from "../../styles/components/blog/Comment.module.scss";
import IconBtn from "../shared/IconBtn";
import User from "../shared/User";
import { AiFillEdit } from "react-icons/ai";
import { MdDelete } from "react-icons/md";
import {
  deleteComment,
  submitComment,
  updateComment,
  voteOnPostComment,
} from "../../services/posts";
import { useEffect, useState } from "react";
import ErrorTip from "../shared/forms/ErrorTip";
import { CommentForm } from "./CommentForm";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";
import { GoChevronUp, GoChevronDown } from "react-icons/go";
import { IComment } from "../../interfaces/PostInterfaces";

export default function Comment({
  comment,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
  updateMyVoteOnComment,
  commentOpened,
  currentParentComment,
}: {
  comment: IComment;
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
  getReplies: (parentId: string) => IComment[];
  updateMyVoteOnComment: (id: string, isUpvote: boolean) => void;
  commentOpened: (to: string) => void;
  currentParentComment: string;
}) {
  const { getUserData } = useUsers();
  const { openModal } = useModal();
  const { user } = useAuth();
  const [err, setErr] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [repliesOpen, setRepliesOpen] = useState(false);
  const childComments = getReplies(comment.ID);

  useEffect(() => {
    if (currentParentComment && currentParentComment === comment.parent_id)
      setRepliesOpen(true);
    // eslint-disable-next-line
  }, [currentParentComment]);

  return (
    <div className={classes.container}>
      <div className={classes.top}>
        <User
          testid="author"
          AdditionalStuff={
            <div className={classes.userIcons}>
              <div className={classes.vert}>
                <div className={classes.votesContainer}>
                  <IconBtn
                    ariaLabel="Vote up"
                    name="Vote up"
                    style={{
                      color: "lime",
                      ...(comment.my_vote && comment.my_vote.is_upvote
                        ? { stroke: "1px" }
                        : { filter: "opacity(0.5)" }),
                    }}
                    svgStyle={{
                      transform: "scale(1.5)",
                      stroke: "var(--text-color)",
                      strokeWidth: "1px",
                    }}
                    Icon={GoChevronUp}
                    onClick={async () => {
                      try {
                        if (!user)
                          throw new Error("You must be logged in to vote");
                        await voteOnPostComment(postId, comment.ID, true);
                        updateMyVoteOnComment(comment.ID, true);
                      } catch (e) {
                        openModal("Message", {
                          err: true,
                          pen: false,
                          msg: `${e}`,
                        });
                      }
                    }}
                    type="button"
                  />
                  {comment.vote_pos_count +
                    (comment.my_vote
                      ? comment.my_vote.is_upvote
                        ? 1
                        : 0
                      : 0) -
                    (comment.vote_neg_count +
                      (comment.my_vote
                        ? comment.my_vote.is_upvote
                          ? 0
                          : 1
                        : 0))}
                  <IconBtn
                    ariaLabel="Vote down"
                    name="Vote down"
                    style={{
                      color: "red",
                      ...(comment.my_vote && !comment.my_vote.is_upvote
                        ? { stroke: "1px" }
                        : { filter: "opacity(0.5)" }),
                    }}
                    svgStyle={{
                      transform: "scale(1.5)",
                      stroke: "var(--text-color)",
                      strokeWidth: "1px",
                    }}
                    Icon={GoChevronDown}
                    onClick={async () => {
                      try {
                        if (!user)
                          throw new Error("You must be logged in to vote");
                        await voteOnPostComment(postId, comment.ID, false);
                        updateMyVoteOnComment(comment.ID, false);
                      } catch (e) {
                        openModal("Message", {
                          err: true,
                          pen: false,
                          msg: `${e}`,
                        });
                      }
                    }}
                    type="button"
                  />
                </div>
              </div>
              {user && user.ID === comment.author_id && (
                <div className={classes.vert}>
                  <IconBtn
                    type="button"
                    ariaLabel="Edit comment"
                    name="Edit comment"
                    onClick={() => setIsEditing(true)}
                    Icon={AiFillEdit}
                  />
                  <IconBtn
                    type="button"
                    ariaLabel="Delete comment"
                    name="Delete comment"
                    style={{ color: "red" }}
                    onClick={() =>
                      openModal("Confirm", {
                        msg: "Are you sure you want to delete this comment?",
                        err: false,
                        pen: false,
                        confirmationCallback: async () => {
                          try {
                            await deleteComment(postId, comment.ID);
                          } catch (e) {
                            setErr(`${e}`);
                          }
                        },
                        cancellationCallback: () => {},
                      })
                    }
                    Icon={MdDelete}
                  />
                </div>
              )}
            </div>
          }
          date={comment.created_at ? new Date(comment.created_at) : new Date()}
          uid={comment.author_id}
          user={getUserData(comment.author_id)}
        />
        <div className={classes.content}>
          {isEditing ? (
            <CommentForm
              autoFocus
              initialValue={comment.content}
              onSubmit={async (c: string) => {
                try {
                  if (!c) throw new Error("You cannot submit an empty comment");
                  if (c.length > 200)
                    throw new Error("Comment length maximum 200 characters");
                  updateComment(postId, comment.ID, c);
                } catch (e) {
                  setErr(`${e}`);
                }
              }}
              onClickOutside={() => setIsEditing(false)}
              ariaLabel="Edit comment input"
            />
          ) : (
            <>{comment.content}</>
          )}
          {err && <ErrorTip message={err} />}
        </div>
        <div className={classes.hor}>
          {user && replyingTo !== comment.ID && (
            <button
              name="Reply to comment"
              aria-label="Reply to comment"
              onClick={() => setReplyingTo(comment.ID)}
              className={classes.showRepliesButtonAndReplyToCommentButton}
            >
              Reply to comment
            </button>
          )}
          {!repliesOpen && childComments && childComments.length > 0 && (
            <button
              onClick={() => {
                setRepliesOpen(true);
                commentOpened(comment.ID);
              }}
              className={classes.showRepliesButtonAndReplyToCommentButton}
            >
              Show replies
            </button>
          )}
        </div>
      </div>
      {replyingTo === comment.ID && (
        <div className={classes.commentForm}>
          <CommentForm
            autoFocus
            onSubmit={async (c: string) => {
              try {
                if (!c) throw new Error("You cannot submit an empty comment");
                if (c.length > 200)
                  throw new Error("Comment length maximum 200 characters");
                submitComment(c, postId, replyingTo);
              } catch (e) {
                openModal("Message", {
                  msg: `${e}`,
                  err: true,
                  pen: false,
                });
              }
            }}
            onClickOutside={() => setReplyingTo("")}
            ariaLabel="Reply to comment input"
          />
        </div>
      )}
      {childComments && repliesOpen && childComments.length > 0 && (
        <div className={classes.childComments}>
          {childComments.map((cmt: IComment) => (
            <Comment
              key={cmt.ID}
              updateMyVoteOnComment={updateMyVoteOnComment}
              getReplies={getReplies}
              setReplyingTo={setReplyingTo}
              replyingTo={replyingTo}
              postId={postId}
              comment={cmt}
              commentOpened={commentOpened}
              currentParentComment={currentParentComment}
            />
          ))}
          {childComments && repliesOpen && (
            <button
              onClick={() => setRepliesOpen(false)}
              className={classes.repliesBar}
            >
              <span />
            </button>
          )}
        </div>
      )}
    </div>
  );
}
