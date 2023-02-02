import { IPostCard } from "../../interfaces/PostInterfaces";
import { memo } from "react";
import classes from "../../styles/components/blog/AsidePostCard.module.scss";

const AsidePostCard = ({ post }: { post?: IPostCard }) => {
  return (
    <article className={classes.container}>
      {post && (
        <a href={`/post/${post.slug}`}>
          <div className={classes.text}>
            <h3>{post.title}</h3>
          </div>
        </a>
      )}
    </article>
  );
};

export default memo(AsidePostCard);
