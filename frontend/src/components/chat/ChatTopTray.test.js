import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { ChatContext, ChatSection } from "../../context/ChatContext";
import ChatTopTray from "./ChatTopTray";

let container = null;

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
});

afterEach(() => {
  unmountComponentAtNode(container);
  container.remove();
  container = null;
});

describe("chat top tray", () => {
  test("should render the close chat icon, and the section", () => {
    render(
      <ChatContext.Provider value={{ section: ChatSection.MENU }}>
        <ChatTopTray />
      </ChatContext.Provider>,
      container
    );

    const closeChatBtn = screen.getByRole("button", {
      name: "Close chat",
      hidden: true,
    });

    expect(closeChatBtn).toBeInTheDocument();
    expect(screen.getByText(ChatSection.MENU)).toBeInTheDocument();
  });

  test("should render the menu button when in a section other than the menu. Clicking on it should bring the user back to the menu", async () => {
    const mockSetSection = jest.fn();

    render(
      <ChatContext.Provider
        value={{ section: ChatSection.ROOMS, setSection: mockSetSection }}
      >
        <ChatTopTray />
      </ChatContext.Provider>,
      container
    );

    const chatMenuBtn = screen.getByRole("button", {
      name: "Chat menu",
      hidden: true,
    });

    await act(async () => {
      chatMenuBtn.click();
    });

    expect(mockSetSection).toHaveBeenCalledWith(ChatSection.MENU);
  });
});
