import classes from "../../styles/components/blog/Comments.module.scss";
import Comment, { IComment } from "./Comment";

/*
 This doesn't really need a test (probably)
*/

export default function Comments({
  comments,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
  updateMyVoteOnComment,
}: {
  comments: IComment[] | null;
  getReplies: (parentId: string) => IComment[];
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
  updateMyVoteOnComment: (id: string, isUpvote: boolean) => void;
}) {
  return (
    <div className={classes.container}>
      {comments &&
        comments.map((cmt) => (
          <Comment
            key={cmt.ID}
            updateMyVoteOnComment={updateMyVoteOnComment}
            getReplies={getReplies}
            setReplyingTo={setReplyingTo}
            replyingTo={replyingTo}
            postId={postId}
            comment={cmt}
          />
        ))}
    </div>
  );
}
