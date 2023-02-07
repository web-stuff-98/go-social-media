import { screen, render, getByTestId } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import Rooms from "./Rooms";

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

describe("chatrooms menu component", () => {
  test("should render the rooms list container, the search controls and the pagination controls", async () => {
    await act(async () => {
      render(<Rooms />, container);
    });

    // getByRole and findByRole not working here again... why???
    const roomsContainer = screen.getByTestId("Rooms list container");
    const roomsResMsgContainer = screen.getByTestId(
      "Rooms list ResMsg container"
    );
    const searchRoomNameInput = screen.getByTestId("Search room name input");
    const searchFormSubmitBtn = screen.getByTestId("Search room button");

    expect(roomsContainer).toBeInTheDocument();
    expect(roomsResMsgContainer).toBeInTheDocument();
    expect(searchRoomNameInput).toBeInTheDocument();
    expect(searchFormSubmitBtn).toBeInTheDocument();

    const paginationControlsContainer = screen.getByTestId(
      "Pagination controls container"
    );
    const paginationControlsNextBtn = screen.getByRole("button", {
      name: "Next page",
    });
    const paginationControlsPrevBtn = screen.getByRole("button", {
      name: "Previous page",
    });

    expect(paginationControlsContainer).toBeInTheDocument();
    expect(paginationControlsNextBtn).toBeInTheDocument();
    expect(paginationControlsPrevBtn).toBeInTheDocument();
  });
});
