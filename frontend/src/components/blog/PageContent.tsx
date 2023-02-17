import { useUsers } from "../../context/UsersContext";
import classes from "../../styles/pages/Page.module.scss";
import IconBtn from "../shared/IconBtn";
import User from "../shared/User";
import { GoChevronUp, GoChevronDown } from "react-icons/go";
import { voteOnPost } from "../../services/posts";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";
import { IPost } from "../../interfaces/PostInterfaces";
import { useInterface } from "../../context/InterfaceContext";
import { useEffect, useState } from "react";

export default function PageContent({
  post,
  setPost,
  imgURL,
}: {
  post: IPost;
  setPost: (to: IPost) => void;
  imgURL: string;
}) {
  const { getUserData } = useUsers();
  const { openModal } = useModal();
  const { user } = useAuth();
  const {
    state: { isMobile },
  } = useInterface();

  const [scrollY, setScrollY] = useState(0);

  const handleScroll = () => setScrollY(window.scrollY);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll);
    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);

  return (
    <>
      <div
        data-testid="Image and title"
        className={classes.imageTitleContainer}
      >
        <div
          className={classes.image}
          style={{
            backgroundImage: `url(${imgURL})`,
            backgroundPositionY: `${Math.floor(scrollY * 0.5)}px`,
          }}
        />
        <div aria-live="assertive" className={classes.text}>
          <div tabIndex={0} className={classes.titleDescription}>
            <h1 data-testid="Heading">{post.title}</h1>
            {!isMobile && <h2 data-testid="Subheading">{post.description}</h2>}
          </div>
          <User
            testid="Author"
            reverse
            light
            date={new Date(post.created_at || 0)}
            uid={post.author_id}
            user={getUserData(post.author_id)}
            AdditionalStuff={
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
                    transform: "scale(1.5)",
                    stroke: "var(--text-color)",
                    strokeWidth: "1px",
                  }}
                  Icon={GoChevronUp}
                  type="button"
                  onClick={async () => {
                    try {
                      if (!user)
                        throw new Error("You must be logged in to vote");
                      await voteOnPost(post.ID, true);
                      setPost({
                        ...post,
                        my_vote: post.my_vote
                          ? null
                          : {
                              uid: user?.ID as string,
                              is_upvote: true,
                            },
                      } as IPost);
                    } catch (e) {
                      openModal("Message", {
                        err: true,
                        pen: false,
                        msg: `${e}`,
                      });
                    }
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
                    transform: "scale(1.5)",
                    stroke: "var(--text-color)",
                    strokeWidth: "1px",
                  }}
                  Icon={GoChevronDown}
                  type="button"
                  onClick={async () => {
                    try {
                      if (!user)
                        throw new Error("You must be logged in to vote");
                      await voteOnPost(post.ID, false);
                      setPost({
                        ...post,
                        my_vote: post.my_vote
                          ? null
                          : {
                              uid: user?.ID as string,
                              is_upvote: false,
                            },
                      } as IPost);
                    } catch (e) {
                      openModal("Message", {
                        err: true,
                        pen: false,
                        msg: `${e}`,
                      });
                    }
                  }}
                />
              </div>
            }
          />
        </div>
      </div>
      <div
        tabIndex={1}
        data-testid="Content"
        className={classes.html}
        dangerouslySetInnerHTML={{ __html: post.body }}
      />
    </>
  );
}
