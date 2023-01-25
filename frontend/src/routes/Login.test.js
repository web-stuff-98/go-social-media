import { render, screen } from "@testing-library/react";
import Login from "./Login";

/*
Useless test. I am learning how to write them.
*/

describe("login form", () => {
  test("should render a login form with a username and password input", () => {
    render(<Login />);

    const usernameInput = screen.getByTestId("username");
    const passwordInput = screen.getByTestId("password");
    const submitButton = screen.getByRole("button");

    expect(usernameInput).toBeInTheDocument;
    expect(passwordInput).toBeInTheDocument;
    expect(submitButton).toBeInTheDocument;
  });
});
