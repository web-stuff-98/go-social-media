import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { AuthContext } from "../context/AuthContext";
import { SocketContext } from "../context/SocketContext";
import { UsersContext } from "../context/UsersContext";
import Page from "./Page";
import * as postServices from "../services/posts";
import { BrowserRouter } from "react-router-dom";
import { act } from "react-dom/test-utils";

let container = null;

const mockUser = {
  ID: "1",
  username: "Test user",
};

jest.mock("react-router-dom", () => ({
  ...jest.requireActual("react-router-dom"),
  useParams: () => ({
    slug: "test-post-slug",
  }),
  useLocation: jest.fn(),
}));

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

describe("post page component", () => {
  test("getPost should be invoked, and afterwards the PageContent component parts should be present, along with the comments and the comment form", async () => {
    postServices.getPost = jest.fn().mockResolvedValue({
      ID: "1",
      author_id: "1",
      title: "Test post title",
      description: "Test post description",
      tags: ["Test tag 1", "Test tag 2"],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      slug: "test-post-slug",
      img_blur: "",
      vote_pos_count: 3,
      vote_neg_count: 3,
      my_vote: null,
      img_url: "",
      body: "<p>Test post body</p>",
    });

    await act(async () => {
      render(
        <SocketContext.Provider
          value={{ openSubscription: jest.fn(), closeSubscription: jest.fn() }}
        >
          <UsersContext.Provider
            value={{ cacheUserData: jest.fn(), getUserData: jest.fn() }}
          >
            <AuthContext.Provider value={{ user: mockUser }}>
              <BrowserRouter>
                <Page />
              </BrowserRouter>
            </AuthContext.Provider>
          </UsersContext.Provider>
        </SocketContext.Provider>,
        container
      );
    });

    expect(postServices.getPost).toHaveBeenCalled();
    expect(screen.getByTestId("Image and title")).toBeInTheDocument();
    expect(screen.getByTestId("Content")).toBeInTheDocument();
    expect(screen.getByTestId("Comments container")).toBeInTheDocument();
    expect(screen.getByTestId("Comment form")).toBeInTheDocument();
  });
});
