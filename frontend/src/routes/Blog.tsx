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
import SearchFormDropdowns from "../components/blog/SearchFormDropdowns";

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
      <div data-testid="Feed container" className={classes.feed}>
        {!resMsg.pen && (
          <>
            {posts.map((p, i) => (
              <PostCard
                reverse={isMobile ? false : Boolean(i % 2)}
                key={p.ID}
                post={p}
              />
            ))}
            <div aria-hidden={true} className={classes.endFix} />
          </>
        )}
        <div style={{ margin: "auto" }}>
          <ResMsg large resMsg={resMsg} />
        </div>
      </div>
      <aside data-testid="Aside container">
        <div className={classes.inner}>
          <form data-testid="Search form" className={classes.searchForm}>
            <SearchFormDropdowns />
            <div className={classes.search}>
              <input
                data-testid="Search input"
                name="Search input"
                id="Search input"
                aria-label="Search"
                type="text"
                value={getTermFromParams}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  setTermInParams(e.target.value)
                }
              />
              <IconBtn
                type="submit"
                name="Search"
                ariaLabel="Search"
                Icon={FaSearch}
              />
            </div>
          </form>
          <div className={classes.searchTags}>
            {getTagsFromParams.map((t) => (
              <button
                aria-label={"Tag:" + t}
                name={"Tag:" + t}
                key={t}
                onClick={() => addOrRemoveTagInParams(t)}
              >
                {t}
              </button>
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
      <div
        data-testid="Pagination controls container"
        className={classes.paginationControls}
      >
        <button
          name="Previous page"
          aria-label="Previous page"
          onClick={() => prevPage()}
        >
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
        <button
          name="Next page"
          aria-label="Next page"
          onClick={() => nextPage()}
        >
          <BsChevronRight />
        </button>
      </div>
    </div>
  );
}
