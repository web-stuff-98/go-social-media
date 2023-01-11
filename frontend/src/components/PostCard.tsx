import { useInterface } from "../context/InterfaceContext";
import { IPostCard } from "../routes/Blog";
import { useState, useEffect, useRef } from "react";
import classes from "../styles/components/PostCard.module.scss";
import { deletePost, getPostThumb } from "../services/posts";
import type { CancelToken, CancelTokenSource } from "axios";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import usePosts from "../context/PostsContext";
import { AiFillEdit } from "react-icons/ai";
import { RiDeleteBin2Fill } from "react-icons/ri";
import IconBtn from "./IconBtn";
import { useModal } from "../context/ModalContext";
import User from "./User";
import { useUsers } from "../context/UsersContext";

export default function PostCard({ post }: { post: IPostCard }) {
  const {
    state: { isMobile },
  } = useInterface();
  const navigate = useNavigate();
  const { postEnteredView, postLeftView } = usePosts();
  const { openModal } = useModal();
  const { getUserData } = useUsers();

  const textContainerRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [imgURL, setImgURL] = useState("");
  const imgCancelSource = useRef<CancelTokenSource>();
  const imgCancelToken = useRef<CancelToken>();

  useEffect(() => {
    postEnteredView(post.ID);
    return () => {
      postLeftView(post.ID);
    };
  }, []);

  //Rerender of image is triggered when img_url query param v=1 increments
  useEffect(() => {
    if (!post.img_url) return;
    imgCancelSource.current = axios.CancelToken.source();
    imgCancelToken.current = imgCancelSource.current?.token;
    getPostThumb(post.ID, imgCancelToken.current)
      .then((url) => setImgURL(url))
      .catch((e) => {
        if (!axios.isCancel(e)) {
          setImgURL("");
          console.error(e);
        }
      });
    return () => {
      if (imgCancelToken.current) {
        imgCancelSource.current?.cancel("Image no longer visible");
      }
    };
  }, [post.img_url]);

  const renderUser = (uid: string) => {
    return (
      <User
        light
        date={new Date(post.created_at || 0)}
        uid={uid}
        user={getUserData(uid)}
      />
    );
  };

  return (
    <article
      ref={containerRef}
      style={isMobile ? { height: "33.33%" } : {}}
      className={classes.container}
    >
      <div
        style={
          isMobile
            ? {}
            : {
                border: "none",
              }
        }
        className={classes.inner}
      >
        {post && (
          <>
            <div
              style={{
                height: `${textContainerRef.current?.clientHeight}px`,
                backgroundImage: `url(${post.img_blur})`,
              }}
              className={classes.imageContainer}
            >
              <img
                onClick={() => navigate(`/post/${post.slug}`)}
                alt={post.title}
                style={isMobile ? { width: "40%", minWidth: "40%" } : {}}
                src={imgURL}
              />
              {
                <div className={classes.userContainer}>
                  {renderUser(post.author_id)}
                </div>
              }
              {
                <div className={classes.actionIcons}>
                  <IconBtn
                    style={{ color: "red", padding: "3px" }}
                    ariaLabel="Delete post"
                    name="Delete post"
                    type="button"
                    onClick={() => {
                      openModal("Confirm", {
                        msg: "Are you sure you want to delete this post?",
                        err: false,
                        pen: false,
                        confirmationCallback: () => {
                          deletePost(post.slug)
                            .then(() =>
                              openModal("Message", {
                                msg: "Post deleted",
                                err: false,
                                pen: false,
                              })
                            )
                            .catch((e) => {
                              openModal("Message", {
                                msg: `${e}`,
                                err: true,
                                pen: false,
                              });
                            });
                        },
                        cancellationCallback: () => {},
                      });
                    }}
                    Icon={RiDeleteBin2Fill}
                  />
                  <IconBtn
                    style={{ color: "white", padding: "3px" }}
                    ariaLabel="Delete post"
                    name="Delete post"
                    type="button"
                    Icon={AiFillEdit}
                    onClick={() => navigate(`/editor/${post.slug}`)}
                  />
                </div>
              }
            </div>
            <div ref={textContainerRef} className={classes.textTags}>
              <h1>{post.title}</h1>
              <h3>{post.description}</h3>
              <div className={classes.tags}>
                {post.tags.map((t) => (
                  <button
                  key={t}
                    aria-label={`Tag ${t}`}
                    id={`Tag ${t}`}
                    name={`Tag ${t}`}
                    className={classes.tag}
                  >
                    {t}
                  </button>
                ))}
              </div>
            </div>
          </>
        )}
      </div>
    </article>
  );
}
