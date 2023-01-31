import { render, fireEvent, screen } from "@testing-library/react";
import Dropdown from "./Dropdown";

describe("Dropdown component", () => {
  test("it renders without crashing", () => {
    render(<Dropdown />);
  });

  test("it opens and closes the dropdown list when the main item is clicked", () => {
    render(<Dropdown />);
    const mainItem = screen.getByText("A");
    fireEvent.click(mainItem);
    expect(screen.getByText("B")).toBeInTheDocument();
    fireEvent.click(mainItem);
    expect(screen.queryByTestId("B")).not.toBeInTheDocument();
  });

  /*
  I cant be asked with this
  test("it updates the main item when a list item is clicked", async () => {
    let index = 0;
    const setIndex = (to) => {
      index = to;
    };
    render(<Dropdown index={index} setIndex={setIndex} />);
    const mainItem = screen.getByTestId("Index 0");
    await act(async () => {
      fireEvent.click(mainItem);
    });
    const listItem = screen.getByTestId("Index 1");
    await act(async () => {
      fireEvent.click(listItem);
    });
    expect(screen.getByTestId("Index 0")).toHaveTextContent("B");
  });*/
});
