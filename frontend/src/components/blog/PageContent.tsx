import { useUsers } from "../../context/UsersContext";
import { IPost } from "../../routes/Page";
import classes from "../../styles/pages/Post.module.scss";
import IconBtn from "../shared/IconBtn";
import User from "../shared/User";
import { TiArrowSortedUp, TiArrowSortedDown } from "react-icons/ti";
import { voteOnPost } from "../../services/posts";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";

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

  return (
    <>
      <div className={classes.imageTitleContainer}>
        <img className={classes.image} alt={post.title} src={imgURL} />
        <div className={classes.text}>
          <div className={classes.titleDescription}>
            <h1 data-testid="heading">{post.title}</h1>
            <h1 data-testid="subheading">{post.description}</h1>
          </div>
          <User
            testid="author"
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
                    transform: "scale(1.166)",
                  }}
                  Icon={TiArrowSortedUp}
                  type="button"
                  onClick={() =>
                    voteOnPost(post.ID, true)
                      .catch((e) => {
                        openModal("Message", {
                          err: true,
                          pen: false,
                          msg: `${e}`,
                        });
                      })
                      .then(() => {
                        setPost({
                          ...post,
                          my_vote: post.my_vote
                            ? null
                            : {
                                uid: user?.ID as string,
                                is_upvote: true,
                              },
                        } as IPost);
                      })
                  }
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
                  Icon={TiArrowSortedDown}
                  type="button"
                  onClick={() =>
                    voteOnPost(post.ID, false)
                      .catch((e) => {
                        openModal("Message", {
                          err: true,
                          pen: false,
                          msg: `${e}`,
                        });
                      })
                      .then(() => {
                        setPost({
                          ...post,
                          my_vote: post.my_vote
                            ? null
                            : {
                                uid: user?.ID as string,
                                is_upvote: false,
                              },
                        } as IPost);
                      })
                  }
                />
              </div>
            }
          />
        </div>
      </div>
      <div
        data-testid="content"
        className={classes.html}
        dangerouslySetInnerHTML={{ __html: post.body }}
      />
    </>
  );
}
