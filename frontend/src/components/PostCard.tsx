import { useInterface } from "../context/InterfaceContext";
import { IPostCard } from "../routes/Blog";
import { useState, useEffect, useRef, useLayoutEffect } from "react";
import classes from "../styles/components/PostCard.module.scss";
import { deletePost, getPostThumb, voteOnPost } from "../services/posts";
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
import { FaChevronUp, FaChevronDown } from "react-icons/fa";
import { useAuth } from "../context/AuthContext";

export default function PostCard({
  post,
  reverse,
}: {
  post: IPostCard;
  reverse: boolean;
}) {
  const {
    state: { isMobile },
  } = useInterface();
  const navigate = useNavigate();
  const {
    postEnteredView,
    postLeftView,
    updatePostCard,
    addOrRemoveTagInParams,
    getTagsFromParams,
  } = usePosts();
  const { openModal } = useModal();
  const { getUserData } = useUsers();
  const { user } = useAuth();

  const textContainerRef = useRef<HTMLDivElement>(null);
  const [imgURL, setImgURL] = useState("");
  const imgCancelSource = useRef<CancelTokenSource>();
  const imgCancelToken = useRef<CancelToken>();
  const containerRef = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  const observer = new IntersectionObserver(([entry]) => {
    if (entry.isIntersecting) {
      setVisible(true);
    } else {
      setVisible(false);
      if (imgCancelSource.current) {
        imgCancelSource.current.cancel();
      }
    }
  });
  useLayoutEffect(() => {
    observer.observe(containerRef.current!);
    return () => {
      observer.disconnect();
    };
    //putting the ref in the dependency array was the only way to get this working properly for some reason
  }, [containerRef.current]);

  useEffect(() => {
    postEnteredView(post.ID);
    return () => {
      postLeftView(post.ID);
    };
  }, []);

  //Rerender of image is triggered when img_url query param v=1 increments
  useEffect(() => {
    if (!post.img_url || !visible) return;
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
  }, [post.img_url, visible]);

  const renderUser = (uid: string) => {
    return (
      <User
        additionalStuff={[
          <div className={classes.votesContainer}>
            <IconBtn
              ariaLabel="Vote up"
              name="Vote up"
              style={{
                color: "lime",
                ...(post.my_vote && post.my_vote.is_upvote
                  ? { stroke: "1px" }
                  : { filter: "opacity(0.5)" }),
              }}
              svgStyle={{
                transform: "scale(1.166)",
              }}
              Icon={FaChevronUp}
              type="button"
              onClick={() => {
                if (user)
                  voteOnPost(post.ID, true)
                    .catch((e) => {
                      openModal("Message", {
                        err: true,
                        pen: false,
                        msg: `${e}`,
                      });
                    })
                    .then(() => {
                      updatePostCard({
                        ID: post.ID,
                        my_vote: post.my_vote
                          ? null
                          : {
                              uid: user?.ID as string,
                              is_upvote: true,
                            },
                      });
                    });
              }}
            />
            {post.vote_pos_count +
              (post.my_vote ? (post.my_vote.is_upvote ? 1 : 0) : 0) -
              (post.vote_neg_count +
                (post.my_vote ? (post.my_vote.is_upvote ? 0 : 1) : 0))}
            <IconBtn
              ariaLabel="Vote down"
              name="Vote down"
              style={{
                color: "red",
                ...(post.my_vote && !post.my_vote.is_upvote
                  ? { stroke: "1px" }
                  : { filter: "opacity(0.5)" }),
              }}
              svgStyle={{
                transform: "scale(1.166)",
              }}
              Icon={FaChevronDown}
              type="button"
              onClick={() => {
                if (user)
                  voteOnPost(post.ID, false)
                    .catch((e) => {
                      openModal("Message", {
                        err: true,
                        pen: false,
                        msg: `${e}`,
                      });
                    })
                    .then(() => {
                      updatePostCard({
                        ID: post.ID,
                        my_vote: post.my_vote
                          ? null
                          : {
                              uid: user?.ID as string,
                              is_upvote: false,
                            },
                      });
                    });
              }}
            />
          </div>,
        ]}
        reverse={reverse}
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
      className={isMobile ? classes.mobileContainer : classes.container}
    >
      <div className={classes.inner}>
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
              {user && user.ID === post.author_id && (
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
              )}
            </div>
            <div ref={textContainerRef} className={classes.textTags}>
              <h1>{post.title}</h1>
              <h3>{post.description}</h3>
              <div className={classes.tags}>
                {post.tags.map((t) => (
                  <button
                    onClick={() => addOrRemoveTagInParams(t)}
                    key={t}
                    aria-label={`Tag: ${t}`}
                    id={`Tag: ${t}`}
                    name={`Tag: ${t}`}
                    className={
                      getTagsFromParams.includes(t)
                        ? classes.tagSelected
                        : classes.tag
                    }
                  >
                    {t}
                  </button>
                ))}
              </div>
              <div
                style={
                  reverse
                    ? { right: "var(--padding)" }
                    : { left: "var(--padding)" }
                }
                className={classes.userContainer}
              >
                {renderUser(post.author_id)}
              </div>
            </div>
          </>
        )}
      </div>
    </article>
  );
}
