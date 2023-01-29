import { render, screen, act, fireEvent } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import Conversations from "./Conversations";
import * as chatServices from "../../services/chat";
import { AuthContext } from "../../context/AuthContext";
import { UsersContext } from "../../context/UsersContext";
import { SocketContext } from "../../context/SocketContext";

let container = null;

const mockMessage = {
  ID: "1",
  uid: "2",
  content: "mock message content",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  has_attachment: false,
  recipient_id: "1",
  attachment_progress: undefined,
  attachment_metadata: undefined,
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

const sendIfPossibleMock = jest.fn();

async function RenderComponent() {
  chatServices.getConversations = jest.fn().mockResolvedValueOnce(["2"]);
  chatServices.getConversation = jest.fn().mockResolvedValueOnce([mockMessage]);
  return await act(async () => {
    render(
      <SocketContext.Provider value={{ sendIfPossible: sendIfPossibleMock }}>
        <AuthContext.Provider value={{ user: mockUser }}>
          <UsersContext.Provider
            value={{ cacheUserData: jest.fn(), getUserData: jest.fn() }}
          >
            <Conversations />
          </UsersContext.Provider>
        </AuthContext.Provider>
      </SocketContext.Provider>,
      container
    );
  });
}

describe("conversations chat section", () => {
  test("should render the conversations section with the user list, the messages list and the message form, getConversations should have been triggered, and the conversations list should be present", async () => {
    await RenderComponent();

    const usersList = screen.getByTestId("Users list");
    const messagesList = screen.getByTestId("Messages and video chat");

    expect(usersList).toBeInTheDocument();
    expect(messagesList).toBeInTheDocument();
    expect(chatServices.getConversations).toHaveReturnedTimes(1);
  });

  test("clicking on a conversation should open up the users messages", async () => {
    await RenderComponent();

    const conversationUserButton = await screen.findByTestId(
      "conversation uid:2"
    );

    expect(conversationUserButton).toBeInTheDocument();

    await act(async () => {
      conversationUserButton.click();
    });

    expect(chatServices.getConversation).toHaveBeenCalled();

    const messageContent = await screen.findByText(mockMessage.content);

    expect(messageContent).toBeInTheDocument();
  });

  test("opening a conversation then filling out the message input and clicking the submit button should invoke the sendIfPossible function", async () => {
    await RenderComponent()

    // getByRole and findByRole wasn't working here for some reason. getByTestId works fine.
    const conversationUserButton = await screen.findByTestId(
      "conversation uid:2"
    );
    await act(async () => {
      conversationUserButton.click();
    });

    const input = screen.getByRole("textbox");
    fireEvent.change(input, {
      target: { value: "Test message" },
    });
    const sendBtn = screen.getByTestId("Send button");

    await act(async () => {
      sendBtn.click();
    });

    expect(sendIfPossibleMock).toHaveBeenCalled();
  });
});
