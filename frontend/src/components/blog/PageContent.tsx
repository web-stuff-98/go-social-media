import { useUsers } from "../../context/UsersContext";
import classes from "../../styles/pages/Page.module.scss";
import IconBtn from "../shared/IconBtn";
import User from "../shared/User";
import { TiArrowSortedUp, TiArrowSortedDown } from "react-icons/ti";
import { voteOnPost } from "../../services/posts";
import { useModal } from "../../context/ModalContext";
import { useAuth } from "../../context/AuthContext";
import { IPost } from "../../interfaces/PostInterfaces";

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
      <div
        data-testid="Image and title"
        className={classes.imageTitleContainer}
      >
        <img className={classes.image} alt={post.title} src={imgURL} />
        <div className={classes.text}>
          <div className={classes.titleDescription}>
            <h1 data-testid="Heading">{post.title}</h1>
            <h1 data-testid="Subheading">{post.description}</h1>
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
                    transform: "scale(1.166)",
                  }}
                  Icon={TiArrowSortedUp}
                  type="button"
                  onClick={async () => {
                    try {
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
                    transform: "scale(1.166)",
                  }}
                  Icon={TiArrowSortedDown}
                  type="button"
                  onClick={async () => {
                    try {
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
        data-testid="Content"
        className={classes.html}
        dangerouslySetInnerHTML={{ __html: post.body }}
      />
    </>
  );
}
