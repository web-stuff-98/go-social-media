import classes from "../../styles/components/Comments.module.scss";
import Comment, { IComment } from "./Comment";

export default function Comments({
  comments,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
}: {
  comments: IComment[] | null;
  getReplies: (parentId: string) => IComment[];
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
}) {
  return (
    <div className={classes.container}>
      {comments &&
        comments.map((cmt) => (
          <Comment
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
