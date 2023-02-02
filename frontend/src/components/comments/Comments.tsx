import { IComment } from "../../interfaces/PostInterfaces";
import classes from "../../styles/components/blog/Comments.module.scss";
import Comment from "./Comment";

/*
 This doesn't really need a test
*/

export default function Comments({
  comments,
  postId,
  replyingTo,
  setReplyingTo,
  getReplies,
  updateMyVoteOnComment,
  commentOpened,
}: {
  comments: IComment[] | null;
  getReplies: (parentId: string) => IComment[];
  postId: string;
  replyingTo: string;
  setReplyingTo: (to: string) => void;
  updateMyVoteOnComment: (id: string, isUpvote: boolean) => void;
  commentOpened: (to: string) => void;
}) {
  /*
  Broken somehow... doesn't give the correct
  count on nested comments... cba to fix
  const getCountOfAllChildren = useCallback(
    (id: string) => {
      let count = 0;
      const replies = getReplies(id);
      if (replies)
        replies.forEach((reply) => {
          count++;
          count += getCountOfAllChildren(reply.ID);
        });
      return count;
    },
    // eslint-disable-next-line
    [comments]
  );*/

  return (
    <div data-testid="Comments container" className={classes.container}>
      {comments &&
        comments.map((cmt) => (
          <Comment
            commentOpened={commentOpened}
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
