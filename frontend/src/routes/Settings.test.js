import { screen, render, fireEvent } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { AuthContext } from "../context/AuthContext";
import { ModalContext, ModalProvider } from "../context/ModalContext";
import Settings from "./Settings";

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

const mockUser = {
  ID: "1",
  username: "Username",
};

describe("settings page", () => {
  test("should render the settings page with the users name, profile picture and the delete account button", () => {
    render(
      <AuthContext.Provider value={{ user: mockUser }}>
        <Settings />
      </AuthContext.Provider>,
      container
    );

    expect(screen.getByTestId("User"));
    expect(screen.getByRole("button", { name: "Delete account" }));
    expect(screen.getByText(mockUser.username));
  });

  test("clicking on the delete account button should open up the confirmation modal. Confirming should call the deleteAccount method from AuthContext", async () => {
    const mockDeleteAccount = jest.fn();

    render(
      <ModalProvider>
        <AuthContext.Provider
          value={{ user: mockUser, deleteAccount: mockDeleteAccount }}
        >
          <Settings />
        </AuthContext.Provider>
      </ModalProvider>,
      container
    );

    const deleteBtn = screen.getByRole("button", { name: "Delete account" });

    await act(async () => {
      deleteBtn.click();
    });

    const modalConfirmBtn = screen.getByRole("button", { name: "Confirm" });

    expect(modalConfirmBtn).toBeInTheDocument();

    await act(async () => {
      modalConfirmBtn.click();
    });

    expect(mockDeleteAccount).toHaveBeenCalled();
  });
});
