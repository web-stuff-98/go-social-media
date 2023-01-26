import { screen, render } from "@testing-library/react";
import { AuthContext } from "../../context/AuthContext";
import Chat from "./Chat";

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
});
