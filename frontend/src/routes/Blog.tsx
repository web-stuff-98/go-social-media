import { useEffect } from "react";
import { BsChevronLeft, BsChevronRight } from "react-icons/bs";
import { useParams, useSearchParams } from "react-router-dom";
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
    is_upvote: boolean;
  };
  img_url: string; //img_url is stored here so that rerender can be triggered when the image is updated by modifying the query string
}

export default function Blog() {
  const { openSubscription, closeSubscription } = useSocket();
  const { resMsg, getPageWithParams, posts, postsCount, nextPage, prevPage } =
    usePosts();
  const { page } = useParams();
  let [searchParams] = useSearchParams();

  useEffect(() => {
    openSubscription("post_feed");
    return () => {
      closeSubscription("post_feed");
    };
  }, []);

  useEffect(() => {
    if (page) {
      getPageWithParams(Number(page));
    }
  }, [page, searchParams]);

  return (
    <div className={classes.container}>
      <div className={classes.feed}>
        {posts.map((p, i) => (
          <PostCard reverse={!Boolean(i % 2)} key={p.ID} post={p} />
        ))}
        <ResMsg resMsg={resMsg} />
      </div>
      <div className={classes.paginationControls}>
        <button onClick={() => prevPage()}>
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
        <button onClick={() => nextPage()}>
          <BsChevronRight />
        </button>
      </div>
    </div>
  );
}
