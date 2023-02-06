import { PostsProvider } from "../context/PostsContext";
import Blog from "./Blog";

export default function BlogWithContext() {
  return (
    <PostsProvider>
      <Blog />
    </PostsProvider>
  );
}
