import { render, screen } from "@testing-library/react";
import NotFound from "./NotFound";

describe("not found page", () => {
  test("should render a page with a heading and a paragraph", () => {
    render(<NotFound />);

    const heading = screen.getByTestId("heading");
    const paragraph = screen.getByTestId("paragraph");

    expect(heading).toBeInTheDocument;
    expect(paragraph).toBeInTheDocument;
  });
});
