import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
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
  test("should render a login form with a username and password input and a submit button. Clicking on the button should send a fetch request.", async () => {
    await act(async () => {
      render(<Login />, container);
    });

    const usernameInput = screen.getByTestId("username");
    const passwordInput = screen.getByTestId("password");
    const submitButton = screen.getByRole("button");

    expect(usernameInput).toBeInTheDocument;
    expect(passwordInput).toBeInTheDocument;
    expect(submitButton).toBeInTheDocument;

    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    await act(async () => {
      submitButton.click();
    });

    expect(axiosSpy).toHaveBeenCalled();

    global.fetch.mockClear();
  });
});
