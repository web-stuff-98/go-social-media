import { screen, render, fireEvent } from "@testing-library/react";
import { act } from "react-dom/test-utils";
import { AuthContext } from "../../context/AuthContext";
import { UsersContext } from "../../context/UsersContext";
import Comment from "./Comment";
import * as postServices from "../../services/posts";
postServices.voteOnPostComment = jest.fn();

let container = null;

const mockComment = {
  ID: "1",
  content: "content",
  author_id: "1",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  parent_id: "",
  vote_pos_count: 3,
  vote_neg_count: 3,
  my_vote: null,
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
  container.remove();
  container = null;
});

describe("comment component", () => {
  test("the comment renders correctly", async () => {
    await act(async () => {
      render(
        <AuthContext.Provider value={{ user: { ID: "1", username: "name" } }}>
          <UsersContext.Provider value={{ getUserData: jest.fn() }}>
            <Comment
              comment={mockComment}
              postId="1"
              replyingTo=""
              setReplyingTo={jest.fn()}
              getReplies={jest.fn()}
              updateMyVoteOnComment={jest.fn()}
            />
          </UsersContext.Provider>
        </AuthContext.Provider>,
        container
      );
    });

    const authorElement = screen.getByTestId("author");
    const voteUpBtn = screen.getByRole("button", {
      name: "Vote up",
    });
    const voteDownBtn = screen.getByRole("button", {
      name: "Vote down",
    });
    const commentContent = screen.getByText(mockComment.content);
    const editBtn = screen.getByRole("button", {
      name: "Edit comment",
    });
    const deleteBtn = screen.getByRole("button", {
      name: "Delete comment",
    });

    expect(authorElement).toBeInTheDocument();
    expect(voteUpBtn).toBeInTheDocument();
    expect(voteDownBtn).toBeInTheDocument();
    expect(commentContent).toBeInTheDocument();
    expect(deleteBtn).toBeInTheDocument();
    expect(editBtn).toBeInTheDocument();
  });

  test("clicking on the voting buttons triggers voteOnPostComment with the correct parameters", async () => {
    await act(async () => {
      render(
        <AuthContext.Provider value={{ user: { ID: "1", username: "name" } }}>
          <UsersContext.Provider value={{ getUserData: jest.fn() }}>
            <Comment
              comment={mockComment}
              postId="1"
              replyingTo=""
              setReplyingTo={jest.fn()}
              getReplies={jest.fn()}
              updateMyVoteOnComment={jest.fn()}
            />
          </UsersContext.Provider>
        </AuthContext.Provider>,
        container
      );
    });

    const voteUpBtn = screen.getByRole("button", {
      name: "Vote up",
    });
    const voteDownBtn = screen.getByRole("button", {
      name: "Vote down",
    });

    voteUpBtn.click();
    voteDownBtn.click();
    expect(postServices.voteOnPostComment).toHaveBeenCalledWith(
      "1",
      mockComment.ID,
      true
    );
    expect(postServices.voteOnPostComment).toHaveBeenCalledWith(
      "1",
      mockComment.ID,
      false
    );
  });
});
