import { useEffect } from "react";
import { ChangeEvent } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { FaSearch } from "react-icons/fa";
import { useParams } from "react-router-dom";
import AsidePostCard from "../components/blog/AsidePostCard";
import Dropdown from "../components/shared/Dropdown";
import IconBtn from "../components/shared/IconBtn";
import PostCard from "../components/blog/PostCard";
import ResMsg from "../components/shared/ResMsg";
import { useInterface } from "../context/InterfaceContext";
import usePosts from "../context/PostsContext";
import useSocket from "../context/SocketContext";
import classes from "../styles/pages/Blog.module.scss";

export interface IPostCard {
  ID: string;
  author_id: string;
  title: string;
  description: string;
  tags: string[];
  created_at: string;
  updated_at: string;
  slug: string;
  img_blur: string;
  vote_pos_count: number; // Excludes users own vote
  vote_neg_count: number; // Excludes users own vote
  my_vote: null | {
    uid: string;
    is_upvote: boolean;
  };
  img_url: string; //img_url is stored here so that rerender can be triggered when the image is updated by modifying the query string
}

export default function Blog() {
  const { openSubscription, closeSubscription } = useSocket();
  const {
    state: { isMobile },
  } = useInterface();
  const {
    resMsg,
    getTagsFromParams,
    getTermFromParams,
    posts,
    postsCount,
    newPosts,
    nextPage,
    prevPage,
    addOrRemoveTagInParams,
    setTermInParams,
    getSortModeFromParams,
    getSortOrderFromParams,
    setSortModeInParams,
    setSortOrderInParams,
  } = usePosts();
  const { page } = useParams();

  useEffect(() => {
    openSubscription("post_feed");
    return () => {
      closeSubscription("post_feed");
    };
    // eslint-disable-next-line
  }, []);

  return (
    <div className={classes.container}>
      <div className={classes.feed}>
        {!resMsg.pen && (
          <>
            {posts.map((p, i) => (
              <PostCard
                reverse={isMobile ? false : Boolean(i % 2)}
                key={p.ID}
                post={p}
              />
            ))}
            <div className={classes.endFix} />
          </>
        )}
        <div style={{ margin: "auto" }}>
          <ResMsg large resMsg={resMsg} />
        </div>
      </div>
      <aside>
        <div className={classes.inner}>
          <form className={classes.searchForm}>
            <div className={classes.dropdownsContainer}>
              <Dropdown
                light
                index={getSortOrderFromParams === "DESC" ? 0 : 1}
                setIndex={setSortOrderInParams}
                items={[
                  { name: "DESC", node: "Desc" },
                  { name: "ASC", node: "Asc" },
                ]}
              />
              <div className={classes.sortMode}>
                <Dropdown
                  light
                  index={getSortModeFromParams === "DATE" ? 0 : 1}
                  setIndex={setSortModeInParams}
                  items={[
                    { name: "DATE", node: "Date" },
                    { name: "POPULARITY", node: "Popularity" },
                  ]}
                />
              </div>
            </div>
            <div className={classes.search}>
              <input
                name="Search input"
                id="Search input"
                aria-label="Search"
                type="text"
                value={getTermFromParams}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  setTermInParams(e.target.value)
                }
                required
              />
              <IconBtn name="Search" ariaLabel="Search" Icon={FaSearch} />
            </div>
          </form>
          <div className={classes.searchTags}>
            {getTagsFromParams.map((t) => (
              <button onClick={() => addOrRemoveTagInParams(t)}>{t}</button>
            ))}
          </div>
          <>
            <h3 className={classes.recentPostsHeading}>New posts</h3>
            <div className={classes.posts}>
              {newPosts.map((p) => (
                <AsidePostCard key={p.ID} post={p} />
              ))}
            </div>
          </>
        </div>
      </aside>
      <div className={classes.paginationControls}>
        <button aria-label="Previous page" onClick={() => prevPage()}>
          <BsChevronLeft />
        </button>
        <div className={classes.text}>
          <span aria-label="Page number">
            {page}/{Math.ceil(postsCount / 20)}
          </span>
          <span aria-label="Number of posts">
            {posts?.length}/{postsCount}
          </span>
        </div>
        <button aria-label="Next page" onClick={() => nextPage()}>
          <BsChevronRight />
        </button>
      </div>
    </div>
  );
}
