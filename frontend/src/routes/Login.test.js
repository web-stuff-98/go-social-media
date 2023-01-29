import { fireEvent, render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { AuthContext } from "../context/AuthContext";
import Login from "./Login";

/*
Useless test. I am learning how to write them.
*/

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

describe("login page", () => {
  test("should render a login form with a username and password input and a submit button", async () => {
    await act(async () => {
      render(
          <Login />,
        container
      );
    });

    const usernameInput = screen.getByTestId("username");
    const passwordInput = screen.getByTestId("password");
    const submitButton = screen.getByRole("button");

    expect(usernameInput).toBeInTheDocument();
    expect(passwordInput).toBeInTheDocument();
    expect(submitButton).toBeInTheDocument();
  });

  test("Inputting a username and password then clicking on the button should trigger the login function from AuthContext", async () => {
    const loginMock = jest.fn();

    await act(async () => {
      render(
        <AuthContext.Provider value={{ login: loginMock }}>
          <Login />
        </AuthContext.Provider>,
        container
      );
    });

    const usernameInput = screen.getByTestId("username");
    const passwordInput = screen.getByTestId("password");
    const submitButton = screen.getByRole("button");

    fireEvent.change(usernameInput, { target: { value: "Test Acc" } });
    fireEvent.change(passwordInput, { target: { value: "Test Pass" } });

    await act(async () => {
      submitButton.click();
    });

    expect(loginMock).toHaveBeenCalledWith("Test Acc", "Test Pass");
  });
});
