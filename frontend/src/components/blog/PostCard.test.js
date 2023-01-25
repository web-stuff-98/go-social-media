import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { BrowserRouter } from "react-router-dom";
import { UsersContext } from "../../context/UsersContext";
import PostCard from "./PostCard";

let container = null;

let upvoteBtn, downvoteBtn;

const getUserData = jest.fn().mockImplementation(() => {
  jest.fn((id) => ({
    id,
    name: "Test User",
  }));
});

const mockPost = {
  ID: "123",
  title: "Test post title",
  description: "Test post description",
  slug: "/post/test-post",
  tags: ["TestTag1", "TestTag2"],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  vote_pos_count: 3,
  vote_neg_count: 3,
  img_blur: "placeholder",
  img_url: "placeholder",
  my_vote: null,
  author_id: "1"
};

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

describe("blog post feed card", () => {
  test("should render the post card inside an article container with the title, description, tags, author and voting buttons", () => {
    render(
      <BrowserRouter>
        <UsersContext.Provider
          value={{ getUserData, users: [{ ID: 1, username: "Test user" }] }}
        >
          <PostCard post={mockPost} />
        </UsersContext.Provider>
      </BrowserRouter>,
      container
    );

    const articleContainer = screen.getByRole("article");
    const title = screen.getByText(mockPost.title);
    const description = screen.getByText(mockPost.description);
    const author = screen.getByTestId("author");
    const expectTags = [];
    mockPost.tags.forEach((t) => {
      expectTags.push(screen.getByText(t));
    });

    upvoteBtn = screen.getByRole("button", {
      name: "Vote up",
      hidden: true,
    });

    downvoteBtn = screen.getByRole("button", {
      name: "Vote down",
      hidden: true,
    });

    expect(articleContainer).toBeInTheDocument;
    expect(title).toBeInTheDocument;
    expect(description).toBeInTheDocument;
    expect(author).toBeInTheDocument;
    expect(upvoteBtn).toBeInTheDocument;
    expect(downvoteBtn).toBeInTheDocument;
    expectTags.forEach((t) => expect(t).toBeInTheDocument);
  });

  test("clicking vote up should make a fetch request", () => {
    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    upvoteBtn.click();

    expect(axiosSpy).toHaveBeenCalled;

    global.fetch.mockClear();
  });

  test("clicking vote down should make a fetch request", () => {
    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    downvoteBtn.click();

    expect(axiosSpy).toHaveBeenCalled;

    global.fetch.mockClear();
  });
});
