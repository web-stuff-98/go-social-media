import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { AuthContext } from "../../context/AuthContext";
import Chat from "./Chat";

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

jest.mock("react-router-dom", () => ({
  ...jest.requireActual("react-router-dom"),
  useLocation: () => ({
    pathname: "doesn't matter",
  }),
}));

describe("the chat component", () => {
  test("should render a chat icon inside an invisible button by default", () => {
    render(
      <AuthContext.Provider value={{ user: { ID: "1" } }}>
        <Chat />
      </AuthContext.Provider>
    );

    const chatButton = screen.getByRole("button", {
      name: "Open chat",
      hidden: true,
    });

    expect(chatButton).toBeInTheDocument();
  });

  test("clicking on the chat icon should open up the menu, with the close chat icon present, clicking on the close chat icon should close the chat", async () => {
    await act(async () => {
      render(
        <AuthContext.Provider value={{ user: { ID: "1" } }}>
          <Chat />
        </AuthContext.Provider>,
        container
      );
    });

    await act(async () => {
      screen
        .getByRole("button", {
          name: "Open chat",
          hidden: true,
        })
        .click();
    });

    const menuContainer = screen.getByTestId("Menu container");
    expect(menuContainer).toBeInTheDocument();
    const closeChatButton = screen.getByRole("button", { name: "Close chat" });
    expect(closeChatButton).toBeInTheDocument();

    await act(async () => {
      closeChatButton.click();
    });

    expect(closeChatButton).not.toBeInTheDocument();
    expect(menuContainer).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: "Open chat",
        hidden: true,
      })
    ).toBeInTheDocument();
  });
});
