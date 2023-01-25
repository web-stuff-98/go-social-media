import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import AsidePostCard from "./AsidePostCard";

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

describe("aside post card component", () => {
  test("should render the title of the post inside of a link", () => {
    render(
      <AsidePostCard
        post={{ title: "Post title", slug: "/page/post-title" }}
      />,
      container
    );

    const linkElement = screen.getByRole("link");
    const titleElement = screen.getByRole("heading");

    expect(linkElement).toBeInTheDocument();
    expect(titleElement).toBeInTheDocument();
  });
});
