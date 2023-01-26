import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import AsidePostCard from "./AsidePostCard";

let container = null;

const mockPost = {
  ID: "123",
  title: "Test post",
};

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
});

afterEach(() => {
  unmountComponentAtNode(container);
  container.remove();
  container = null;
});

describe("the aside post card component for the recent posts feed", () => {
  test("should render the post card in an article element with the heading inside of an anchor element", () => {
    render(<AsidePostCard post={mockPost} />, container);

    const articleContainer = screen.getByRole("article");
    const link = screen.getByRole("link");
    const heading = screen.getByText(mockPost.title);

    expect(articleContainer).toBeInTheDocument();
    expect(link).toBeInTheDocument();
    expect(heading).toBeInTheDocument();
  });
});
