import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import RoomMessage from "./RoomMessage";

const mockMessage = {
  ID: 1,
  content: "Hello",
  has_attachment: true,
  attachment_metadata: {
    type: "application/pdf",
    name: "test.pdf",
    size: 4 * 1024 * 1024,
  },
  attachment_progress: {
    failed: false,
    pending: false,
    ratio: 1,
  },
  created_at: "2023-01-01T12:00:00Z",
  updated_at: "2023-01-01T12:00:00Z",
};

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

describe("room message component", () => {
  test("renders the message content", () => {
    render(<RoomMessage msg={mockMessage} reverse={false} />, container);
    expect(screen.getByText("Hello")).toBeInTheDocument();
  });

  test("renders the date and time of the message", async () => {
    await act(async () => {
      render(<RoomMessage msg={mockMessage} reverse={false} />, container);
    });
    expect(screen.getByTestId("Date")).toHaveTextContent("01/01/23");
    expect(screen.getByTestId("Time")).toHaveTextContent("12:00");
  });

  test("renders the message in reverse", () => {
    render(<RoomMessage msg={mockMessage} reverse={true} />, container);
    expect(screen.getByTestId("Container")).toHaveStyle(
      "flex-direction: row-reverse"
    );

    const contentContainer = screen.getByTestId("Content container");
    expect(contentContainer).toHaveStyle("textAlign: right");

    const dateContainer = screen.getByTestId("Date container");
    expect(dateContainer).toHaveStyle("textAlign: left");
  });
});
