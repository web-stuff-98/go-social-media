import { useEffect } from "react";
import { ChangeEvent } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { FaSearch } from "react-icons/fa";
import { useParams, useSearchParams } from "react-router-dom";
import Dropdown from "../components/Dropdown";
import IconBtn from "../components/IconBtn";
import PostCard from "../components/PostCard";
import ResMsg from "../components/ResMsg";
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
    resMsg,
    getTagsFromParams,
    getTermFromParams,
    posts,
    postsCount,
    nextPage,
    prevPage,
    getPageWithParams,
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
    getPageWithParams()
    return () => {
      closeSubscription("post_feed");
    };
  }, []);

  return (
    <div className={classes.container}>
      <div className={classes.feed}>
        {!resMsg.pen && (
          <>
            {posts.map((p, i) => (
              <PostCard reverse={false} key={p.ID} post={p} />
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
                index={getSortOrderFromParams() === "DESC" ? 0 : 1}
                setIndex={setSortOrderInParams}
                items={[
                  { name: "DESC", node: "Desc" },
                  { name: "ASC", node: "Asc" },
                ]}
              />
              <Dropdown
                light
                index={getSortModeFromParams() === "DATE" ? 0 : 1}
                setIndex={setSortModeInParams}
                items={[
                  { name: "DATE", node: "Date" },
                  { name: "POPULARITY", node: "Popularity" },
                ]}
              />
            </div>
            <input
              name="Search input"
              id="Search input"
              aria-label="Search"
              type="text"
              value={getTermFromParams()}
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                setTermInParams(e.target.value)
              }
              required
            />
            <IconBtn name="Search" ariaLabel="Search" Icon={FaSearch} />
          </form>
          <div className={classes.searchTags}>
            {getTagsFromParams().map((t) => (
              <button onClick={() => addOrRemoveTagInParams(t)}>{t}</button>
            ))}
          </div>
        </div>
      </aside>
      <div className={classes.paginationControls}>
        <button onClick={() => prevPage()}>
          <BsChevronLeft />
        </button>
        <div className={classes.text}>
          <span aria-label="Page number">
            {page}/{Math.ceil(postsCount / 30)}
          </span>
          <span aria-label="Number of posts">
            {posts?.length}/{postsCount}
          </span>
        </div>
        <button onClick={() => nextPage()}>
          <BsChevronRight />
        </button>
      </div>
    </div>
  );
}
