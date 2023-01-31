import classes from "../styles/pages/Page.module.scss";
import { useEffect, useState, useCallback, useMemo } from "react";
import { getPost, submitComment } from "../services/posts";
import { useNavigate, useParams } from "react-router-dom";
import ResMsg from "../components/shared/ResMsg";
import useSocket from "../context/SocketContext";
import {
  instanceOfChangeData,
  instanceOfPostCommentVoteData,
  instanceOfPostVoteData,
} from "../utils/DetermineSocketEvent";
import { baseURL } from "../services/makeRequest";
import { CommentForm } from "../components/comments/CommentForm";
import Comments from "../components/comments/Comments";
import { useUsers } from "../context/UsersContext";
import { useAuth } from "../context/AuthContext";
import PageContent from "../components/blog/PageContent";
import { IComment, IPost } from "../interfaces/PostInterfaces";
import { IResMsg } from "../interfaces/GeneralInterfaces";

export default function Page() {
  const { openSubscription, closeSubscription, socket } = useSocket();
  const { slug } = useParams();
  const { cacheUserData } = useUsers();
  const { user } = useAuth();
  const navigate = useNavigate();

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });
  const [post, setPost] = useState<IPost>();
  const [imgURL, setImgURL] = useState("");
  const [replyingTo, setReplyingTo] = useState("");
  const [comments, setComments] = useState<IComment[]>([]);
  const commentsByParentId = useMemo<any>(() => {
    const group: any = {};
    comments.forEach((cmt) => {
      group[cmt.parent_id] ||= [];
      group[cmt.parent_id].push(cmt);
    });
    return group;
  }, [comments]);
  const [parentComment, setParentComment] = useState<string | null>("");
  const getReplies = (parentId: string): IComment[] =>
    commentsByParentId[parentId as keyof typeof commentsByParentId];

  const loadPost = async () => {
    if (!slug) return;
    setResMsg({ msg: "", err: false, pen: true });
    try {
      const p = await getPost(slug);
      setPost({
        ...(p as Omit<IPost, "comments">),
        my_vote:
          p.my_vote?.uid === "000000000000000000000000" ? null : p.my_vote,
      });
      setComments(
        p.comments
          ? p.comments.map((cmt: IComment) => {
              let outCmt = cmt;
              if (cmt.my_vote?.uid === "000000000000000000000000")
                outCmt.my_vote = null;
              return outCmt;
            })
          : []
      );
      setResMsg({ msg: "", err: false, pen: false });
      cacheUserData(p.author_id);
      setImgURL(`${baseURL}/api/posts/${p.ID}/image?v=1`);
      openSubscription(`post_page=${p.ID}`);
    } catch (e) {
      setResMsg({ msg: `${e}`, err: true, pen: false });
    }
  };

  useEffect(() => {
    loadPost();
    return () => {
      closeSubscription(`post_page=${post?.ID}`);
    };
    // eslint-disable-next-line
  }, [slug]);

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "POST") {
        if (data.DATA.ID !== post?.ID) return;
        if (data.METHOD === "DELETE") {
          navigate("/blog/1");
          return;
        }
        if (data.METHOD === "UPDATE") {
          setPost((o) => ({ ...o, ...data.DATA } as IPost));
          return;
        }
        if (data.METHOD === "UPDATE_IMAGE") {
          setImgURL(`${baseURL}/api/posts/${post.ID}/image?v=${Math.random()}`);
          return;
        }
      }
      if (data.ENTITY === "POST_COMMENT") {
        if (data.METHOD === "INSERT") {
          //@ts-ignore
          cacheUserData(data.DATA.author_id);
          setComments((o) => [...o, data.DATA as IComment]);
          return;
        }
        if (data.METHOD === "DELETE") {
          setComments((o) => [
            ...o.filter(
              (c) => c.ID !== data.DATA.ID || c.parent_id === data.DATA.ID
            ),
          ]);
          return;
        }
        if (data.METHOD === "UPDATE") {
          setComments((o) => {
            let newCmts = o;
            const i = o.findIndex((c) => c.ID === data.DATA.ID);
            if (i === -1) return o;
            newCmts[i] = { ...newCmts[i], ...(data.DATA as Partial<IComment>) };
            return newCmts;
          });
          return;
        }
      }
    }
    if (instanceOfPostVoteData(data)) {
      setPost((p) => {
        let newPost = p;
        if (!newPost) return;
        if (data.DATA.is_upvote) {
          newPost.vote_pos_count += data.DATA.remove ? -1 : 1;
        } else {
          newPost.vote_neg_count += data.DATA.remove ? -1 : 1;
        }
        return { ...newPost };
      });
      return;
    }
    if (instanceOfPostCommentVoteData(data)) {
      setComments((cmts) => {
        let newCmts = cmts;
        const i = cmts.findIndex((c) => c.ID === data.DATA.ID);
        if (i === -1) return cmts;
        if (data.DATA.remove) {
          if (data.DATA.is_upvote) {
            newCmts[i].vote_pos_count--;
          } else {
            newCmts[i].vote_neg_count--;
          }
        } else {
          if (data.DATA.is_upvote) {
            newCmts[i].vote_pos_count++;
          } else {
            newCmts[i].vote_neg_count++;
          }
        }
        return [...newCmts];
      });
      return;
    }
    // eslint-disable-next-line
  }, []);

  useEffect(() => {
    if (socket) socket?.addEventListener("message", handleMessage);
    return () => {
      if (socket) socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  const [cmtErr, setCmtErr] = useState("");

  const updateMyVoteOnComment = (id: string, isUpvote: boolean) => {
    setComments((cmts) => {
      let newCmts = cmts;
      const i = cmts.findIndex((cmt) => cmt.ID === id);
      if (i === -1) return cmts;
      newCmts[i].my_vote = newCmts[i].my_vote
        ? null
        : {
            uid: user?.ID as string,
            is_upvote: isUpvote,
          };
      return [...newCmts];
    });
  };

  return (
    <div className={classes.container}>
      {post && <PageContent post={post} imgURL={imgURL} setPost={setPost} />}
      <div className={classes.comments}>
        <CommentForm
          loading={false}
          error={cmtErr}
          onSubmit={async (c: string) => {
            try {
              await submitComment(c, post!.ID, "");
            } catch (e) {
              setCmtErr(`${e}`);
            }
          }}
          placeholder="Add a comment..."
        />
        <Comments
          setReplyingTo={setReplyingTo}
          replyingTo={replyingTo}
          getReplies={getReplies}
          postId={post?.ID as string}
          comments={commentsByParentId[parentComment as string]}
          updateMyVoteOnComment={updateMyVoteOnComment}
        />
      </div>
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
