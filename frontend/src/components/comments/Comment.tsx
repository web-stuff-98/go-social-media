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
import { useState } from "react";
import { FaReply } from "react-icons/fa";
import ErrorTip from "../shared/forms/ErrorTip";
import { CommentForm } from "./CommentForm";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";

import { TiArrowSortedUp, TiArrowSortedDown } from "react-icons/ti";

export interface IComment {
  ID: string;
  content: string;
  author_id: string;
  created_at: string;
  updated_at: string;
  parent_id: string;
  vote_pos_count: number; // Excludes users own vote
  vote_neg_count: number; // Excludes users own vote
  my_vote: null | {
    uid: string;
    is_upvote: boolean;
  };
}

export default function Comment({
  comment,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
  updateMyVoteOnComment,
}: {
  comment: IComment;
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
  getReplies: (parentId: string) => IComment[];
  updateMyVoteOnComment: (id: string, isUpvote: boolean) => void;
}) {
  const { getUserData } = useUsers();
  const { openModal } = useModal();
  const { user } = useAuth();
  const [err, setErr] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [repliesOpen, setRepliesOpen] = useState(false);
  const childComments = getReplies(comment.ID);

  return (
    <div className={classes.container}>
      <div className={classes.top}>
        <User
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
                    Icon={TiArrowSortedUp}
                    onClick={() => {
                      if (user)
                        voteOnPostComment(postId, comment.ID, true)
                          .catch((e) => {
                            openModal("Message", {
                              err: true,
                              pen: false,
                              msg: `${e}`,
                            });
                          })
                          .then(() => {
                            updateMyVoteOnComment(comment.ID, true);
                          });
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
                    Icon={TiArrowSortedDown}
                    onClick={() => {
                      if (user)
                        voteOnPostComment(postId, comment.ID, false)
                          .catch((e) => {
                            openModal("Message", {
                              err: true,
                              pen: false,
                              msg: `${e}`,
                            });
                          })
                          .then(() => {
                            updateMyVoteOnComment(comment.ID, false);
                          });
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
                    name="edit"
                    onClick={() => setIsEditing(true)}
                    Icon={AiFillEdit}
                  />
                  <IconBtn
                    type="button"
                    ariaLabel="Delete comment"
                    name="delete"
                    style={{ color: "red" }}
                    onClick={() =>
                      openModal("Confirm", {
                        msg: "Are you sure you want to delete this comment?",
                        err: false,
                        pen: false,
                        confirmationCallback: () =>
                          deleteComment(postId, comment.ID).catch((e) =>
                            setErr(`${e}`)
                          ),
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
        {user && (
          <div className={classes.icons}>
            <IconBtn
              onClick={() => setReplyingTo(comment.ID)}
              Icon={FaReply}
              name="Reply to comment..."
              ariaLabel="Reply to comment..."
            />
          </div>
        )}
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
      {childComments && repliesOpen && childComments.length > 0 && (
        <div className={classes.childComments}>
          {childComments.map((cmt: IComment) => (
            <Comment
              updateMyVoteOnComment={updateMyVoteOnComment}
              getReplies={getReplies}
              setReplyingTo={setReplyingTo}
              replyingTo={replyingTo}
              postId={postId}
              comment={cmt}
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
