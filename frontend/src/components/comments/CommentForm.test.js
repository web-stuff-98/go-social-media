import { render, fireEvent, screen } from "@testing-library/react";
import "@testing-library/jest-dom/extend-expect";
import { CommentForm } from "./CommentForm";
import { act } from "react-dom/test-utils";

let container = null;

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
});

afterEach(() => {
  container.remove();
  container = null;
});

describe("comment form", () => {
  test("submits the form with the message input", async () => {
    const onSubmit = jest.fn();
    render(<CommentForm onSubmit={onSubmit} />, container);
    const form = screen.getByTestId("form");
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "test message" } });
    fireEvent.submit(form);
    expect(onSubmit).toHaveBeenCalledWith("test message");
  });

  test("calls onClickOutside when clicked outside", async () => {
    const onClickOutside = jest.fn();
    render(
      <CommentForm onSubmit={() => {}} onClickOutside={onClickOutside} />,
      container
    );
    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    await act(async () => {
      fireEvent.mouseEnter(input);
      fireEvent.mouseLeave(input);
      fireEvent.mouseDown(input);
    });
    expect(onClickOutside).toHaveBeenCalled();
  });
});
