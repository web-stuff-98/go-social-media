import {
  render,
  screen,
  act,
  fireEvent,
  waitFor,
} from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { AuthContext } from "../../context/AuthContext";
import { UsersContext } from "../../context/UsersContext";
import { SocketContext } from "../../context/SocketContext";
import * as roomServices from "../../services/rooms";
import Room from "./Room";

let container = null;

const mockMessage = {
  ID: "1",
  uid: "2",
  content: "mock message content",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  has_attachment: false,
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
  roomServices.getRoom = jest.fn().mockResolvedValueOnce({
    ID: "1",
    name: "Test room",
    author_id: "1",
    img_blur: "",
    img_url: "",
    messages: [mockMessage],
  });
  return await act(async () => {
    render(
      <SocketContext.Provider
        value={{
          sendIfPossible: sendIfPossibleMock,
          openSubscription: jest.fn(),
          closeSubscription: jest.fn(),
        }}
      >
        <AuthContext.Provider value={{ user: mockUser }}>
          <UsersContext.Provider
            value={{ cacheUserData: jest.fn(), getUserData: jest.fn() }}
          >
            <Room />
          </UsersContext.Provider>
        </AuthContext.Provider>
      </SocketContext.Provider>,
      container
    );
  });
}

describe("room chat section", () => {
  test("should render the room section with the messages list and the message form, getRoom should have been triggered, and a message should be present", async () => {
    await RenderComponent();

    const messagesList = screen.getByTestId("Messages and videochat");
    const messageForm = screen.getByTestId("Message form");

    expect(messagesList).toBeInTheDocument();
    expect(messageForm).toBeInTheDocument();
    expect(roomServices.getRoom).toHaveReturnedTimes(1);

    expect(screen.getByText(mockMessage.content)).toBeInTheDocument();
  });

  test("filling out the message input and clicking the send button should invoke the sendIfPossible function", async () => {
    await RenderComponent();

    const input = screen.getByRole("textbox");
    fireEvent.change(input, {
      target: { value: "Test message" },
    });
    const sendBtn = screen.getByTestId("Send message");

    await act(async () => {
      sendBtn.click();
    });

    expect(sendIfPossibleMock).toHaveBeenCalled();
  });
});
