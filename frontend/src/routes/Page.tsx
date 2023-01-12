import classes from "../styles/pages/Page.module.scss";
import { IPostCard } from "./Blog";
import { useEffect, useState, useCallback, useMemo } from "react";
import { getPost, submitComment } from "../services/posts";
import { useNavigate, useParams } from "react-router-dom";
import ResMsg, { IResMsg } from "../components/ResMsg";
import useSocket from "../context/SocketContext";
import {
  instanceOfChangeData,
  instanceOfPostVoteData,
} from "../utils/DetermineSocketEvent";
import { baseURL } from "../services/makeRequest";
import { CommentForm } from "../components/comments/CommentForm";
import { IComment } from "../components/comments/Comment";
import Comments from "../components/comments/Comments";
import { useUsers } from "../context/UsersContext";

export interface IPost extends IPostCard {
  body: string;
}

export default function Page() {
  const { openSubscription, closeSubscription, socket } = useSocket();
  const { slug } = useParams();
  const { cacheUserData } = useUsers();
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

  useEffect(() => {
    if (!slug) return;
    setResMsg({ msg: "", err: false, pen: true });
    getPost(slug)
      .then((p) => {
        setPost({
          ...(p as Omit<IPost, "comments">),
          my_vote:
            p.my_vote?.uid === "000000000000000000000000" ? null : p.my_vote,
        });
        setComments(p.comments || []);
        setResMsg({ msg: "", err: false, pen: false });
        setImgURL(`${baseURL}/api/posts/${p.ID}/image?v=1`);
        openSubscription(`post_page=${p.ID}`);
      })
      .catch((e) => setResMsg({ msg: `${e}`, err: true, pen: false }));
    return () => {
      closeSubscription(`post_page=${post?.ID}`);
    };
  }, [slug]);

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    console.log(data);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "POST") {
        if (data.DATA.ID !== post?.ID) return;
        if (data.METHOD === "DELETE") {
          navigate("/blog/1");
        }
        if (data.METHOD === "UPDATE") {
          setPost((o) => ({ ...o, ...data.DATA } as IPost));
        }
        if (data.METHOD === "UPDATE_IMAGE") {
          setImgURL(
            `${baseURL}/api/posts/${post.ID}/image?v=` + `${Math.random()}`
          );
        }
      }
      if (data.ENTITY === "POST_COMMENT") {
        if (data.METHOD === "INSERT") {
          //@ts-ignore
          cacheUserData(data.DATA.author_id);
          setComments((o) => [...o, data.DATA as IComment]);
        }
        if (data.METHOD === "DELETE") {
          setComments((o) => [...o.filter((c) => c.ID !== data.DATA.ID)]);
        }
        if (data.METHOD === "UPDATE") {
          setComments((o) => {
            let newCmts = o;
            const i = o.findIndex((c) => c.ID === data.DATA.ID);
            if (i === -1) return o;
            newCmts[i] = { ...newCmts[i], ...(data.DATA as Partial<IComment>) };
            return newCmts;
          });
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
    }
  }, []);

  useEffect(() => {
    if (socket) socket?.addEventListener("message", handleMessage);
    return () => {
      if (socket) socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  const [cmtErr, setCmtErr] = useState("");

  return (
    <div className={classes.container}>
      {post && (
        <>
          <div
            style={{ backgroundImage: `url(${post.img_blur})` }}
            className={classes.imageTitleContainer}
          >
            <img className={classes.image} alt={post.title} src={imgURL} />
            <div className={classes.text}>
              <div className={classes.titleDescription}>
                <h1>{post.title}</h1>
                <h2>{post.description}</h2>
              </div>
            </div>
          </div>
          <div
            className={classes.html}
            dangerouslySetInnerHTML={{ __html: post.body }}
          />
        </>
      )}
      <div className={classes.comments}>
        <CommentForm
          loading={false}
          error={cmtErr}
          onSubmit={(c: string) =>
            submitComment(c, post!.ID, "").catch((e) => setCmtErr(`${e}`))
          }
          placeholder="Add a comment..."
        />
        <Comments
          setReplyingTo={setReplyingTo}
          replyingTo={replyingTo}
          getReplies={getReplies}
          postId={post?.ID as string}
          comments={commentsByParentId[parentComment as string]}
        />
      </div>
      <ResMsg resMsg={resMsg} />
    </div>
  );
}
