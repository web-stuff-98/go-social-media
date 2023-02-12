import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { PostsContext } from "../context/PostsContext";
import Blog from "./Blog";

let container = null;

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
  const mockIntersectionObserver = jest.fn();
  mockIntersectionObserver.mockReturnValue({
    observe: () => null,
    unobserve: () => null,
    disconnect: () => null,
  });
  window.IntersectionObserver = mockIntersectionObserver;
});

afterEach(() => {
  unmountComponentAtNode(container);
  container.remove();
  container = null;
});

jest.mock("react-router-dom", () => ({
  ...jest.requireActual("react-router-dom"),
  useLocation: () => ({
    pathname: "doesn't matter",
  }),
  useNavigate: jest.fn(),
}));

const mockPost = {
  ID: "1",
  author_id: "1",
  title: "Test post title",
  description: "Test post description",
  tags: ["Test tag 1", "Test tag 2"],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  slug: "test-post-slug",
  img_blur: "placeholder",
  vote_pos_count: 3,
  vote_neg_count: 3,
  my_vote: null,
  img_url: "placeholder",
};

describe("blog page", () => {
  test("should render the pagination controls, the aside container and the search & sort controls. The ResMsg loading spinner should also be present.", async () => {
    await act(async () => {
      render(
        <PostsContext.Provider
          value={{
            resMsg: { msg: "", err: false, pen: true },
            posts: [],
            newPosts: [],
            getPageWithParams: [],
            getTagsFromParams: [],
            getSortModeFromParams: "DATE",
            getSortOrderFromParams: "ASCENDING",
            postEnteredView: jest.fn(),
            postLeftView: jest.fn(),
          }}
        >
          <Blog />
        </PostsContext.Provider>,
        container
      );
    });

    //Containers
    expect(screen.getByTestId("Aside container")).toBeInTheDocument();
    expect(screen.getByTestId("Feed container")).toBeInTheDocument();
    expect(
      screen.getByTestId("Pagination controls container")
    ).toBeInTheDocument();

    //Pagination
    expect(
      screen.getByRole("button", { name: "Previous page", hidden: true })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Next page", hidden: true })
    ).toBeInTheDocument();

    //Search & sort controls
    expect(screen.getByTestId("Search form")).toBeInTheDocument();
    expect(screen.getByTestId("Search input")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Search" })).toBeInTheDocument();
    expect(screen.getByTestId("Search form dropdowns")).toBeInTheDocument();

    //Spinner
    expect(screen.getByTestId("Loading spinner")).toBeInTheDocument();
  });

  test("should render a post inside of the post feed, and inside of the aside feed", async () => {
    render(
      <PostsContext.Provider
        value={{
          resMsg: { msg: "", err: false, pen: false },
          posts: [mockPost],
          newPosts: [{...mockPost, title:"Test post title two"}],
          getPageWithParams: [],
          getTagsFromParams: [],
          getSortModeFromParams: "DATE",
          getSortOrderFromParams: "ASCENDING",
        }}
      >
        <Blog />
      </PostsContext.Provider>
    );

    expect(await screen.findByText("Test post title")).toBeInTheDocument();
    expect(await screen.findByText("Test post title two")).toBeInTheDocument();
  });
});
