import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { BrowserRouter } from "react-router-dom";
import { UsersContext } from "../../context/UsersContext";
import PageContent from "./PageContent";

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
  test("should render the blog post heading, subheading, HTML content, author & voting buttons", async () => {
    render(
      <BrowserRouter>
        <UsersContext.Provider
          value={{
            getUserData,
            userEnteredView,
            userLeftView,
            users: [{ ID: "1", username: "Test user" }],
          }}
        >
          <PageContent post={mockPost} setPost={jest.fn()} />
        </UsersContext.Provider>
      </BrowserRouter>,
      container
    );

    const heading = screen.getByTestId("heading");
    const subheading = screen.getByTestId("subheading");
    const content = screen.getByTestId("content");
    const author = screen.getByTestId("author");

    upvoteBtn = screen.getByRole("button", {
      name: "Vote up",
      hidden: true,
    });
    downvoteBtn = screen.getByRole("button", {
      name: "Vote down",
      hidden: true,
    });

    expect(heading).toBeInTheDocument;
    expect(subheading).toBeInTheDocument;
    expect(content).toBeInTheDocument;
    expect(author).toBeInTheDocument;
    expect(upvoteBtn).toBeInTheDocument;
    expect(downvoteBtn).toBeInTheDocument;
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
