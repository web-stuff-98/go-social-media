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

describe("room message component", () => {
  test("renders the message content", () => {
    render(<RoomMessage msg={mockMessage} reverse={false} />, container);
    expect(screen.getByText("Hello")).toBeInTheDocument();
  });

  test("renders the message in reverse", () => {
    render(<RoomMessage msg={mockMessage} reverse={true} />, container);
    expect(screen.getByTestId("Message container")).toHaveStyle(
      "flex-direction: row-reverse"
    );

    const contentContainer = screen.getByTestId("Text container");
    expect(contentContainer).toHaveStyle("textAlign: right");
  });
});
