import { render, screen } from "@testing-library/react";
import Register from "./Register";

/*
Useless test. I am learning how to write them.
*/

describe("registration form", () => {
  test("should render a registration form with a username and password input", () => {
    render(<Register />);

    const usernameInput = screen.getByTestId("username");
    const passwordInput = screen.getByTestId("password");
    const submitButton = screen.getByRole("button");

    expect(usernameInput).toBeInTheDocument;
    expect(passwordInput).toBeInTheDocument;
    expect(submitButton).toBeInTheDocument;
  });
});
