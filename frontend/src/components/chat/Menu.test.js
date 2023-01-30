import { screen, render, waitFor } from "@testing-library/react";
import { act } from "react-dom/test-utils";
import { BrowserRouter } from "react-router-dom";
import Chat, { ChatContext } from "./Chat";
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

jest.mock("react-router-dom", () => ({
  ...jest.requireActual("react-router-dom"),
  useLocation: () => ({
    pathname: "doesn't matter",
  }),
}));

async function RenderComponent() {
  //Render the chat component instead of the Menu component. Menu component should be inside Chat component.
  await act(async () => {
    render(
      <ChatContext.Provider value={{ setSection: jest.fn() }}>
        <Chat />
      </ChatContext.Provider>,
      container
    );
  });

  //Open the chat so that the menu component is actually present
  const openChatBtn = screen.getByRole("button", { name: "Open chat" });
  await act(async () => {
    openChatBtn.click();
  });
}

describe("chat menu section", () => {
  test("should render the buttons for conversations, rooms and the room editor", async () => {
    await RenderComponent();

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

  test("clicking on the conversations menu button should open conversations. the room menu icon should be present", async () => {
    await RenderComponent();

    const conversationsMenuButton = screen.getByRole("button", {
      name: "Conversations",
    });

    await act(async () => {
      conversationsMenuButton.click();
    });

    expect(screen.getByText("Conversations")).toBeInTheDocument();
  });

  test("clicking on the rooms menu button should open the chatrooms menu. the room menu icon should be present", async () => {
    await RenderComponent();

    const chatroomsMenuButton = screen.getByRole("button", {
      name: "Rooms",
    });

    await act(async () => {
      chatroomsMenuButton.click();
    });

    expect(screen.getByText("Rooms")).toBeInTheDocument();
  });

  test("clicking on the room editor menu button should open the room editor. the room menu icon should be present", async () => {
    await RenderComponent();

    const roomEditorMenuButton = screen.getByRole("button", {
      name: "Editor",
    });

    await act(async () => {
      roomEditorMenuButton.click();
    });

    expect(screen.getByText("Editor")).toBeInTheDocument();
  });
});
