import { render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import Editor from "./Editor";
import * as postServices from "../services/posts";
import { BrowserRouter } from "react-router-dom";

postServices.createPost = jest.fn().mockReturnValueOnce("test-post-slug");
postServices.updatePost = jest.fn().mockReturnValueOnce("test-post-slug");
postServices.getRandomImage = jest.fn();

/*
Useless test. I am learning how to write them.
*/

let container,
  titleInput,
  descriptionInput,
  tagsInput,
  quillContainer,
  quillEditor,
  submitButton,
  selectImageButton,
  randomImageButton;

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
  test("should render a form with a title, description, tags, and HTML editor input with a submit button, a select image button and a random image button", async () => {
    await act(async () => {
      render(
        <BrowserRouter>
          <Editor />
        </BrowserRouter>,
        container
      );
    });

    titleInput = screen.getByTestId("title");
    descriptionInput = screen.getByTestId("description");
    tagsInput = screen.getByTestId("tags");
    quillContainer = screen.getByTestId("quill container");
    quillEditor = document.getElementsByClassName("ql-editor ql-blank");

    submitButton = screen.getByRole("button", {
      name: "Submit",
    });
    selectImageButton = screen.getByTestId("Image file button");
    randomImageButton = screen.getByTestId("Random image button");

    expect(titleInput).toBeInTheDocument();
    expect(descriptionInput).toBeInTheDocument();
    expect(tagsInput).toBeInTheDocument();
    expect(quillContainer).toBeInTheDocument();
    expect(quillEditor[0]).toBeInTheDocument();

    expect(submitButton).toBeInTheDocument();
    expect(selectImageButton).toBeInTheDocument();
    expect(randomImageButton).toBeInTheDocument();
  });
});
