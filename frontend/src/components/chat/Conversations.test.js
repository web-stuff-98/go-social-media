import { render, screen, act, waitFor } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import Conversations from "./Conversations";
import * as chatServices from "../../services/chat";
import { AuthContext } from "../../context/AuthContext";
import { UsersContext } from "../../context/UsersContext";

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

describe("conversations chat section", () => {
  test("should render the conversations section with the user list, the messages list and the message form. getConversations should have been triggered", async () => {
    //list of uids
    chatServices.getConversations = jest.fn().mockResolvedValueOnce(["2", "3"]);
    chatServices.getConversation = jest.fn().mockResolvedValueOnce([
      {
        ID: "1",
        uid: "2",
        content: "content",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        has_attachment: false,
        recipient_id: "1",
        attachment_progress: undefined,
        attachment_metadata: undefined,
      },
    ]);

    await act(async () => {
      render(
        <AuthContext.Provider
          value={{ user: { ID: "1", username: "username" } }}
        >
          <UsersContext.Provider
            value={{ cacheUserData: jest.fn(), getUserData: jest.fn() }}
          >
            <Conversations />
          </UsersContext.Provider>
        </AuthContext.Provider>,
        container
      );
    });

    const usersList = screen.getByTestId("Users list");
    const messagesList = screen.getByTestId("Messages and video chat");

    expect(usersList).toBeInTheDocument();
    expect(messagesList).toBeInTheDocument();
    expect(chatServices.getConversations).toHaveReturnedTimes(1);

    let conversationUserButton

    await act(async () => {
      conversationUserButton = screen.getByTestId("conversation uid:2");
    });

    expect(conversationUserButton).toBeInTheDocument();

    await act(async () => {
      await conversationUserButton.click();
    });

    expect(chatServices.getConversation).toHaveBeenCalled();
  });
});
