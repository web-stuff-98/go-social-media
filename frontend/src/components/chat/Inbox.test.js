import { render, screen, act, fireEvent } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import Inbox from "./Inbox";
import * as chatServices from "../../services/chat";
import { AuthContext } from "../../context/AuthContext";
import { UsersContext } from "../../context/UsersContext";
import { SocketContext } from "../../context/SocketContext";
import React from "react";

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
  invitation: false,
  invitation_accepted: false,
  invitation_declined: false,
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
    const mockChildMethod = jest.fn();
    jest.spyOn(React, "useRef").mockReturnValue({
      current: {
        childMethod: mockChildMethod,
      },
    });
    render(
      <SocketContext.Provider value={{ sendIfPossible: sendIfPossibleMock }}>
        <AuthContext.Provider value={{ user: mockUser }}>
          <UsersContext.Provider
            value={{ cacheUserData: jest.fn(), getUserData: jest.fn() }}
          >
            <Inbox />
          </UsersContext.Provider>
        </AuthContext.Provider>
      </SocketContext.Provider>,
      container
    );
  });
}

describe("inbox chat section", () => {
  test("should render the inbox section with the user list, the messages list and the message form, getConversations should have been triggered, and the conversations list should be present", async () => {
    const mockChildMethod = jest.fn();
    jest.spyOn(React, "useRef").mockReturnValue({
      current: {
        childMethod: mockChildMethod,
      },
    });
    await RenderComponent();

    const usersList = screen.getByTestId("Users list");
    const messagesList = screen.getByTestId("Messages and video chat");
    const messageForm = screen.getByTestId("Message form");

    expect(usersList).toBeInTheDocument();
    expect(messagesList).toBeInTheDocument();
    expect(messageForm).toBeInTheDocument();
    expect(chatServices.getConversations).toHaveReturnedTimes(1);
  });

  test("clicking on a conversation should open up the users messages", async () => {
    const mockChildMethod = jest.fn();
    jest.spyOn(React, "useRef").mockReturnValue({
      current: {
        childMethod: mockChildMethod,
      },
    });
    await RenderComponent();

    const conversationUserButton = await screen.findByTestId(
      "conversation uid:2"
    );

    expect(conversationUserButton).toBeInTheDocument();

    await act(async () => {
      conversationUserButton.click();
    });

    expect(chatServices.getConversation).toHaveBeenCalled();
    expect(screen.getByText(mockMessage.content)).toBeInTheDocument();
  });

  test("opening a conversation then filling out the message input and clicking the send button should invoke the sendIfPossible function", async () => {
    const mockChildMethod = jest.fn();
    jest.spyOn(React, "useRef").mockReturnValue({
      current: {
        childMethod: mockChildMethod,
      },
    });
    await RenderComponent();

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
