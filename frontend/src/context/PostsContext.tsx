import {
  useState,
  useContext,
  createContext,
  useEffect,
  useCallback,
  useMemo,
  startTransition,
  useRef,
} from "react";
import type { ReactNode } from "react";
import {
  getNewestPosts,
  getPage,
  SortMode,
  SortOrder,
} from "../services/posts";
import useSocket from "./SocketContext";
import {
  instanceOfChangeData,
  instanceOfPostVoteData,
} from "../utils/DetermineSocketEvent";
import { baseURL } from "../services/makeRequest";
import { useUsers } from "./UsersContext";
import {
  useLocation,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import debounce from "lodash/debounce";
import { IPostCard } from "../interfaces/PostInterfaces";
import { IResMsg } from "../interfaces/GeneralInterfaces";
import axios, { CancelToken, CancelTokenSource } from "axios";

export const PostsContext = createContext<{
  posts: IPostCard[];
  postsCount: number;
  newPosts: IPostCard[];

  postEnteredView: (id: string) => void;
  postLeftView: (id: string) => void;

  getPageWithParams: () => void;

  updatePostCard: (data: Partial<IPostCard>) => void;
  removePostCard: (id: string) => void;

  getSortOrderFromParams: SortOrder;
  getSortModeFromParams: SortMode;
  getTagsFromParams: string[];
  getTermFromParams: string;
  addOrRemoveTagInParams: (tag: string) => void;

  nextPage: () => void;
  prevPage: () => void;

  setSortOrderInParams: (to: number) => void;
  setSortModeInParams: (to: number) => void;
  setTermInParams: (to: string) => void;

  resMsg: IResMsg;
}>({
  posts: [],
  postsCount: 0,
  newPosts: [],

  postEnteredView: () => {},
  postLeftView: () => {},

  getPageWithParams: () => {},

  updatePostCard: () => {},
  removePostCard: () => {},

  getSortOrderFromParams: "DESCENDING",
  getSortModeFromParams: "DATE",
  getTagsFromParams: [],
  getTermFromParams: "",
  addOrRemoveTagInParams: () => {},

  setSortOrderInParams: () => {},
  setSortModeInParams: () => {},
  setTermInParams: () => {},

  nextPage: () => {},
  prevPage: () => {},

  resMsg: { msg: "", err: false, pen: false },
});

export const PostsProvider = ({ children }: { children: ReactNode }) => {
  const { socket, openSubscription, closeSubscription } = useSocket();
  const { cacheUserData } = useUsers();
  const navigate = useNavigate();
  const { page } = useParams();
  const { search: queryString } = useLocation();
  let [searchParams] = useSearchParams();

  //if value is an empty string, it removes the param from the url
  const addUpdateOrRemoveParamsAndNavigateToUrl = (
    name?: string,
    value?: string
  ) => {
    const rawTags =
      name === "tags" ? value : searchParams.get("tags")?.replaceAll(" ", "+");
    const rawTerm =
      name === "term" ? value : searchParams.get("term")?.replaceAll(" ", "+");
    const rawOrder = name === "order" ? value : searchParams.get("order");
    const rawMode = name === "mode" ? value : searchParams.get("mode");
    let outTags = rawTags ? `&tags=${rawTags}` : "";
    let outTerm = rawTerm ? `&term=${rawTerm}` : "";
    let outOrder = rawOrder ? `&order=${rawOrder}` : "";
    let outMode = rawMode ? `&mode=${rawMode}` : "";
    navigate(
      `/blog/1${outTags}${outTerm}${outOrder.toLowerCase()}${outMode.toLowerCase()}`.replace(
        "/blog/1&",
        "/blog/1?"
      )
    );
  };

  const getSortOrderFromParams = useMemo(
    () => (searchParams.get("order") || "DESCENDING") as SortOrder,
    [searchParams]
  );
  const getSortModeFromParams = useMemo(
    () => (searchParams.get("mode") || "DATE") as SortMode,
    [searchParams]
  );
  const getTagsFromParams = useMemo(
    () =>
      searchParams.has("tags")
        ? searchParams
            .get("tags")!
            .split(" ")
            .filter((t) => t)
        : [],
    [searchParams]
  );
  const getTermFromParams = useMemo(
    () => searchParams.get("term") || "",
    [searchParams]
  );

  const setSortOrderInParams = (index: number) =>
    addUpdateOrRemoveParamsAndNavigateToUrl(
      "order",
      index === 0 ? "DESCENDING" : "ASCENDING"
    );
  const setSortModeInParams = (index: number) =>
    addUpdateOrRemoveParamsAndNavigateToUrl(
      "mode",
      index === 0 ? "DATE" : "POPULARITY"
    );
  const addOrRemoveTagInParams = (tag: string) =>
    addUpdateOrRemoveParamsAndNavigateToUrl(
      "tags",
      (getTagsFromParams.includes(tag)
        ? getTagsFromParams.filter((t) => t !== tag).sort()
        : [...getTagsFromParams, tag].sort()
      ).join("+")
    );
  const setTermInParams = (term: string) =>
    addUpdateOrRemoveParamsAndNavigateToUrl("term", term.replaceAll(" ", "+"));

  const [newPosts, setNewPosts] = useState<IPostCard[]>([]);
  const [posts, setPosts] = useState<IPostCard[]>([]);
  const [postsCount, setPostsCount] = useState(0); // Document count of all posts matching query... not the count of posts on the page
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const postEnteredView = (id: string) => {
    openSubscription(`post_card=${id}`);
  };

  const postLeftView = (id: string) => {
    closeSubscription(`post_card=${id}`);
  };

  const handleSearch = useMemo(
    () =>
      debounce(() => {
        if (page) {
          getPageWithParams();
        }
      }, 300),
    // eslint-disable-next-line
    [searchParams, page]
  );

  useEffect(
    () => handleSearch(),
    // eslint-disable-next-line
    [page, searchParams]
  );

  const cancelSource = useRef<CancelTokenSource>();
  const cancelToken = useRef<CancelToken>();

  const getPageWithParams = () => {
    setResMsg({ msg: "", err: false, pen: true });
    if (cancelSource.current) {
      cancelSource.current.cancel();
    }
    cancelSource.current = axios.CancelToken.source();
    cancelToken.current = cancelSource.current?.token;
    getPage(
      page ? Number(page) : 1,
      getSortOrderFromParams,
      getSortModeFromParams,
      getTagsFromParams,
      getTermFromParams,
      cancelToken.current!
    )
      .then((p: any) => {
        const posts = JSON.parse(p.posts) || [];
        startTransition(() => {
          setPosts(
            posts.map((p: IPostCard) => ({
              ...p,
              img_url: `${baseURL}/api/posts/${p.ID}/image?v=1`,
              my_vote:
                p.my_vote?.uid === "000000000000000000000000"
                  ? null
                  : p.my_vote,
            }))
          );
          setPostsCount(Number(p.count));
        });
        posts.forEach((p: IPostCard) => cacheUserData(p.author_id));
        setResMsg({ msg: "", err: false, pen: false });
      })
      .catch((e) => {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      });
    getNewestPosts().then((p: IPostCard[] | null) => {
      startTransition(() => {
        setNewPosts(p || []);
      });
    });
  };

  const updatePostCard = (data: Partial<IPostCard>) => {
    startTransition(() => {
      setPosts((o) => {
        let newPosts = o;
        const i = o.findIndex((p) => p.ID === data.ID);
        if (i === -1) return o;
        newPosts[i] = { ...newPosts[i], ...data };
        return [...newPosts];
      });
    });
  };

  const updatePostCardImage = (data: { ID: string }) => {
    startTransition(() => {
      setPosts((o) => {
        let newPosts = o;
        const i = o.findIndex((p) => p.ID === data.ID);
        if (i === -1) return o;
        const v = Number(newPosts[i].img_url.split("?v=")[1]);
        newPosts[i] = {
          ...newPosts[i],
          ...data,
          img_url: `${newPosts[i].img_url.split("?v=")[0] + (v + 1)}`,
        };
        return [...newPosts];
      });
    });
  };
  const removePostCard = (id: string) => {
    startTransition(() => {
      setPosts((o) => [...o.filter((p) => p.ID !== id)]);
    });
  };

  const createPostCard = (data: IPostCard, addToStart?: boolean) => {
    startTransition(() => {
      if (addToStart) setPosts((o) => [data, ...o].slice(0, 20));
      else setPosts((o) => [...o, data].slice(0, 20));
    });
  };

  const createPostCardOnNewest = (data: IPostCard) => {
    startTransition(() => {
      setNewPosts((p) => [data, ...p].slice(0, 15));
    });
  };

  const removePostCardFromNewest = (id: string) => {
    startTransition(() => {
      setNewPosts((p) => [...p.filter((p) => p.ID !== id)]);
    });
  };

  const updatePostCardOnNewest = (data: IPostCard) => {
    startTransition(() => {
      setNewPosts((p) => {
        let newPosts = p;
        const i = p.findIndex((p) => p.ID === data.ID);
        if (i === -1) return p;
        newPosts[i] = { ...newPosts[i], ...data };
        return [...newPosts];
      });
    });
  };

  const isPostOnPage = (id: string) =>
    Boolean(posts.findIndex((p) => p.ID === id));
  const isPostOnNewestFeed = (id: string) =>
    Boolean(newPosts.findIndex((p) => p.ID === id));

  const handleMessage = useCallback((e: MessageEvent) => {
    if (!page) return;
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfChangeData(data)) {
      if (data.ENTITY === "POST") {
        if (data.METHOD === "DELETE") {
          if (isPostOnPage(data.DATA.ID)) removePostCard(data.DATA.ID);
          if (isPostOnNewestFeed(data.DATA.ID))
            removePostCardFromNewest(data.DATA.ID);
          return;
        }
        if (data.METHOD === "UPDATE") {
          if (isPostOnPage(data.DATA.ID)) updatePostCard(data.DATA);
          if (isPostOnNewestFeed(data.DATA.ID))
            updatePostCardOnNewest(data.DATA as IPostCard);
          return;
        }
        if (data.METHOD === "UPDATE_IMAGE") {
          if (isPostOnPage(data.DATA.ID))
            updatePostCardImage({ ID: data.DATA.ID });
          return;
        }
        if (data.METHOD === "INSERT") {
          // Add new posts to new posts array
          createPostCardOnNewest(data.DATA as IPostCard);
          setPostsCount((p) => p + 1);
          if (getSortModeFromParams === "DATE") {
            if (getSortOrderFromParams === "DESCENDING" && page === "1") {
              /* If sorting by most recent posts and
              on the first page, add the post to the top
          of the feed */
              //@ts-ignore
              cacheUserData(String(data.DATA.author_id));
              createPostCard(data.DATA as IPostCard, true);
              if (posts[0]) {
                removePostCard(posts[0].ID);
              }
            } else if (
              getSortOrderFromParams === "ASCENDING" &&
              posts.length < 20
            ) {
              /* If sorting by oldest posts and there are less than 20
          posts (the maximum number of posts on a page) it means the
          user is on the last page, so add the post to the end of
          the list */
              //@ts-ignore
              cacheUserData(String(data.DATA.author_id));
              createPostCard(data.DATA as IPostCard);
            } else {
              // Otherwise just refresh the page...
              getPageWithParams();
            }
          } else {
            // Otherwise just refresh the page...
            getPageWithParams();
          }
        }
      }
    }
    if (instanceOfPostVoteData(data)) {
      if (isPostOnPage(data.DATA.ID)) {
        startTransition(() => {
          setPosts((p) => {
            let newPosts = p;
            const i = p.findIndex((p) => p.ID === data.DATA.ID);
            if (i === -1) return p;
            if (data.DATA.is_upvote) {
              newPosts[i].vote_pos_count += data.DATA.remove ? -1 : 1;
            } else {
              newPosts[i].vote_neg_count += data.DATA.remove ? -1 : 1;
            }
            return [...newPosts];
          });
        });
      }
    }
    // eslint-disable-next-line
  }, []);

  const nextPage = () =>
    navigate(
      "/blog/" +
        Math.min(Number(page) + 1, Math.ceil(postsCount / 20)) +
        queryString
    );

  const prevPage = () =>
    navigate("/blog/" + Math.max(Number(page) - 1, 1) + queryString);

  useEffect(() => {
    if (socket) socket?.addEventListener("message", handleMessage);
    return () => {
      if (socket) socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  return (
    <PostsContext.Provider
      value={{
        posts,
        postsCount,
        newPosts,
        postEnteredView,
        postLeftView,
        getPageWithParams,
        updatePostCard,
        removePostCard,
        setSortModeInParams,
        setSortOrderInParams,
        setTermInParams,
        getTagsFromParams,
        getSortOrderFromParams,
        getSortModeFromParams,
        getTermFromParams,
        resMsg,
        nextPage,
        prevPage,
        addOrRemoveTagInParams,
      }}
    >
      {children}
    </PostsContext.Provider>
  );
};

const usePosts = () => useContext(PostsContext);

export default usePosts;
