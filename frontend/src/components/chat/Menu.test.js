import { screen, render } from "@testing-library/react";
import { act } from "react-dom/test-utils";
import { ChatContext } from "./Chat";
import Menu from "./Menu";

let container = null;

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
});

afterEach(() => {
  container.remove();
  container = null;
});

describe("chat menu section", () => {
  test("should render the buttons for conversations, rooms and the room editor", async () => {
    await act(async () => {
      render(
        <ChatContext.Provider value={{ setSection: jest.fn() }}>
          <Menu />
        </ChatContext.Provider>,
        container
      );
    });

    const conversationsMenuButton = screen.getByRole("button", {
      name: "Conversations",
    });
    const chatroomsMenuButton = screen.getByRole("button", {
      name: "Rooms",
    });
    const roomEditorMenuButton = screen.getByRole("button", {
      name: "Editor",
    });

    expect(conversationsMenuButton).toBeInTheDocument();
    expect(chatroomsMenuButton).toBeInTheDocument();
    expect(roomEditorMenuButton).toBeInTheDocument();
  });
});
