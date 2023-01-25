import { render, screen } from "@testing-library/react";
import NotFound from "./NotFound";

/*
Useless test. I am learning how to write them.
*/

describe("not found page", () => {
  test("should render a page with a heading and a paragraph", () => {
    render(<NotFound />);

    const heading = screen.getByTestId("heading");
    const paragraph = screen.getByTestId("paragraph");

    expect(heading).toBeInTheDocument;
    expect(paragraph).toBeInTheDocument;
  });
});
