import { render, screen, act } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { BrowserRouter } from "react-router-dom";
import { UsersContext } from "../../context/UsersContext";
import PageContent from "./PageContent";
import * as postServices from "../../services/posts";
import { AuthContext } from "../../context/AuthContext";

postServices.voteOnPost = jest.fn();

let container = null;
let upvoteBtn, downvoteBtn;

const mockPost = {
  ID: "123",
  title: "Test title",
  description: "Test description",
  tags: ["TestTag1", "TestTag2"],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  slug: "test-title",
  img_blur: "",
  vote_pos_count: 3,
  vote_neg_count: 3,
  my_vote: null,
  img_url: "",
  body: "<p>Test body</p>",
  author_id: "1",
};

const mockUser = { ID: "1", username: "username" };

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

describe("post page content", () => {
  const getUserData = jest.fn().mockImplementation(() => {
    jest.fn((id) => ({
      id,
      name: "Test User",
    }));
  });
  const userEnteredView = jest.fn().mockImplementation((uid) => {});
  const userLeftView = jest.fn().mockImplementation((uid) => {});

  test("should render the blog post heading, subheading, HTML content, author & voting buttons. Clicking on the voting buttons should trigger voteOnPost", async () => {
    await act(async () =>
      render(
        <BrowserRouter>
        <AuthContext.Provider value={{user:mockUser}}>
          <UsersContext.Provider
            value={{
              getUserData,
              userEnteredView,
              userLeftView,
              users: [{ ID: "1", username: "Test user" }],
            }}
          >
            <PageContent
              post={mockPost}
              voteOnPost={jest.fn()}
              setPost={jest.fn()}
            />
          </UsersContext.Provider>
          </AuthContext.Provider>
        </BrowserRouter>,
        container
      )
    );

    const heading = screen.getByTestId("Heading");
    const subheading = screen.getByTestId("Subheading");
    const content = screen.getByTestId("Content");
    const author = screen.getByTestId("Author");

    upvoteBtn = screen.getByRole("button", {
      name: "Vote up",
      hidden: true,
    });
    downvoteBtn = screen.getByRole("button", {
      name: "Vote down",
      hidden: true,
    });

    expect(heading).toBeInTheDocument();
    expect(subheading).toBeInTheDocument();
    expect(content).toBeInTheDocument();
    expect(author).toBeInTheDocument();
    expect(upvoteBtn).toBeInTheDocument();
    expect(downvoteBtn).toBeInTheDocument();

    await act(async () => {
      upvoteBtn.click();
    });
    expect(postServices.voteOnPost).toHaveBeenCalledWith("123", true);

    await act(async () => {
      downvoteBtn.click();
    });
    expect(postServices.voteOnPost).toHaveBeenCalledWith("123", false);
  });
});
