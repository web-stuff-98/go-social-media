import { render, screen, act } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { BrowserRouter } from "react-router-dom";
import { UsersContext } from "../../context/UsersContext";
import PageContent from "./PageContent";

import { voteOnPost } from "../../services/posts";

let container = null;

const mockSetPost = jest.fn();
jest.mock("../../services/posts", () => {
  return {
    voteOnPost: jest.fn(() => new Promise((resolve) => setTimeout(resolve, 0))),
  };
});

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
    const mockData = {
      title: "Test title",
      description: "Test description",
      tags: ["TestTagA", "TestTagB"],
      created_at: "Tue Jan 24 2023 19:32:54 GMT+0000 (Greenwich Mean Time)",
      updated_at: "Tue Jan 24 2023 19:32:54 GMT+0000 (Greenwich Mean Time)",
      slug: "test-title",
      img_blur: "",
      vote_pos_count: 3,
      vote_neg_count: 3,
      my_vote: null,
      img_url: "",
      body: "Test body",
    };

    const getUserData = jest.fn().mockImplementation(() => {
      jest.fn((id) => ({
        id,
        name: "Test User",
      }));
    });

    render(
      <BrowserRouter>
        <UsersContext.Provider
          value={{ getUserData, users: [{ ID: "1", username: "Test User" }] }}
        >
          <PageContent post={mockData} voteOnPost={voteOnPost} setPost={jest.fn()} />
        </UsersContext.Provider>
      </BrowserRouter>,
      container
    );

    const heading = screen.getByTestId("heading");
    const subheading = screen.getByTestId("subheading");
    const content = screen.getByTestId("content");
    const author = screen.getByTestId("author");
    const upvoteBtn = screen.getByRole("button", {
      name: "Vote up",
      hidden: true,
    });
    const downvoteBtn = screen.getByRole("button", {
      name: "Vote down",
      hidden: true,
    });

    expect(heading).toBeInTheDocument;
    expect(subheading).toBeInTheDocument;
    expect(content).toBeInTheDocument;
    expect(author).toBeInTheDocument;
    expect(upvoteBtn).toBeInTheDocument;
    expect(downvoteBtn).toBeInTheDocument;

    const voteOnPostSpy = jest.spyOn(voteOnPost);
    voteOnPostSpy.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 0))
    );

    upvoteBtn.click();
    expect(voteOnPost).toHaveBeenCalledWith(mockData.ID, true);

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0));
    });

    expect(mockSetPost).toHaveBeenCalledWith({
      ...mockData,
      my_vote: {
        uid: "1",
        is_upvote: true,
      },
    });

    voteOnPostSpy.mockReset();
    voteOnPost.mockRestore();
  });
});
