import { useAuth } from "../context/AuthContext";
import { IPostCard } from "../routes/Blog";
import classes from "../styles/components/AsidePostCard.module.scss";

export default function AsidePostCard({ post }: { post?: IPostCard }) {
  return (
    <article className={classes.container}>
      {post && (
        <a href={`/post/${post.slug}`}>
          <div className={classes.text}>
            <h3>{post.title}</h3>
            <p>{post.description}</p>
          </div>
        </a>
      )}
    </article>
  );
}
