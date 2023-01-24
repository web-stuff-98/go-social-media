import { render, screen } from "@testing-library/react";
import Login from "./Login";

test("should render a login form with a username and password input", () => {
  render(<Login />);

  const usernameInput = screen.getByTestId("username");
  const passwordInput = screen.getByTestId("password");

  expect(usernameInput).toBeInTheDocument();
  expect(passwordInput).toBeInTheDocument();
});
