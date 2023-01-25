import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import Editor from "./Editor";

/*
Useless test. I am learning how to write them.
*/

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

describe("blog post editor", () => {
  test("should render a form with a title, description, tags, and HTML editor input with a submit button, a select image button and a random image button. Clicking on the submit button and the random image button should send fetch requests.", async () => {
    await act(async () => {
      render(<Editor />, container);
    });

    const titleInput = screen.getByTestId("title");
    const descriptionInput = screen.getByTestId("description");
    const tagsInput = screen.getByTestId("tags");
    const quillContainer = screen.getByTestId("quill container");

    const submitButton = screen.getByRole("button", {
      name: "Submit",
    });
    const selectImageButton = screen.getByTestId("Image file button");
    const randomImageButton = screen.getByTestId("Random image button");

    expect(titleInput).toBeInTheDocument;
    expect(descriptionInput).toBeInTheDocument;
    expect(tagsInput).toBeInTheDocument;
    expect(quillContainer).toBeInTheDocument;

    expect(submitButton).toBeInTheDocument;
    expect(selectImageButton).toBeInTheDocument;
    expect(randomImageButton).toBeInTheDocument;

    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    await act(async () => {
      submitButton.click();
    });

    await act(async () => {
      randomImageButton.click();
    });

    expect(axiosSpy).toHaveBeenCalledTimes;

    global.fetch.mockClear();
  });
});
