import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { BrowserRouter } from "react-router-dom";
import { UsersContext } from "../../context/UsersContext";
import PostCard from "./PostCard";
import * as postServices from "../../services/posts";
import { AuthContext } from "../../context/AuthContext";

postServices.voteOnPost = jest.fn();

let container = null;
let upvoteBtn, downvoteBtn;

const getUserData = jest.fn().mockImplementation(() => {
  jest.fn((id) => ({
    id,
    name: "Test User",
  }));
});

const userEnteredView = jest.fn().mockImplementation((uid) => {});
const userLeftView = jest.fn().mockImplementation((uid) => {});

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
  body: "<p>Test body</p>",
  my_vote: null,
  author_id: "1",
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
        <AuthContext.Provider value={{ user: { ID: "1" } }}>
          <UsersContext.Provider
            value={{
              getUserData,
              userEnteredView,
              userLeftView,
              users: [{ ID: "1", username: "Test user" }],
            }}
          >
            <PostCard post={mockPost} />
          </UsersContext.Provider>
        </AuthContext.Provider>
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

    expect(articleContainer).toBeInTheDocument();
    expect(title).toBeInTheDocument();
    expect(description).toBeInTheDocument();
    expect(author).toBeInTheDocument();
    expect(upvoteBtn).toBeInTheDocument();
    expect(downvoteBtn).toBeInTheDocument();
    expectTags.forEach((t) => expect(t).toBeInTheDocument());

    upvoteBtn.click();
    expect(postServices.voteOnPost).toHaveBeenCalledWith("123", true);

    downvoteBtn.click();
    expect(postServices.voteOnPost).toHaveBeenCalledWith("123", false);
  });
});
