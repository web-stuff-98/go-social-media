import { fireEvent, render, screen } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import Editor from "./Editor";
import * as postServices from "../services/posts";
import { BrowserRouter } from "react-router-dom";

postServices.updatePost = jest.fn().mockReturnValueOnce("test-post-slug");
postServices.getPostImageFile = jest.fn().mockResolvedValueOnce(
  new File(
    [
      Buffer.from(
        "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"
      ),
    ],
    "image.jpg",
    {
      type: "image/jpeg",
    }
  )
);
postServices.getPost = jest.fn().mockResolvedValueOnce({
  ID: "1",
  title: "Test post title",
  description: "Test post description",
  body: "<p>Test post content</p>",
  tags: ["tag one", "tag two"],
  vote_pos_count: 3,
  vote_neg_count: 3,
  my_vote: null,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  author_id: "1",
  slug: "test-post-slug",
  img_url: "",
  img_blur: "",
});


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

jest.mock("react-router-dom", () => ({
  ...jest.requireActual("react-router-dom"),
  useParams: () => ({
    slug: "test-post-slug",
  }),
}));

describe("blog post editor update post", () => {
  test("should render a form with a title, description, tags, and HTML editor input with an update button, a select image button and a random image button", async () => {
    await act(async () => {
      render(<Editor />, container);
    });

    titleInput = screen.getByTestId("title");
    descriptionInput = screen.getByTestId("description");
    tagsInput = screen.getByTestId("tags");
    quillContainer = screen.getByTestId("quill container");
    quillEditor = document.getElementsByClassName("ql-editor ql-blank");

    submitButton = screen.getByText("Update");
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

  test("when there is a slug in the url params, that means that the editor is editing a post. All the inputs should be in the document. getPost and getPostImageFile should be triggered, and filling out the inputs and clicking the submit button should trigger updatePost", async () => {
    await act(async () => {
      render(
        <BrowserRouter>
          <Editor />
        </BrowserRouter>,
        container
      );
    });

    expect(postServices.getPost).toHaveBeenCalledWith("test-post-slug");

    titleInput = screen.getByTestId("title");
    descriptionInput = screen.getByTestId("description");
    tagsInput = screen.getByTestId("tags");
    quillEditor = document.getElementsByClassName("ql-editor ql-blank");

    fireEvent.change(titleInput, {
      target: { value: "Test title placeholder" },
    });
    fireEvent.change(descriptionInput, {
      target: { value: "Test description placeholder" },
    });
    fireEvent.change(tagsInput, {
      target: { value: "#Test tag 1 #Test tag 2" },
    });

    await act(async () => {
      quillEditor[0].innerHTML = "<p>Test quill content</p>";
    });

    submitButton = screen.getByText("Update");

    await act(async () => {
      await submitButton.click();
    });

    expect(postServices.updatePost).toHaveBeenCalledWith(
      {
        title: "Test title placeholder",
        description: "Test description placeholder",
        tags: "#Test tag 1 #Test tag 2",
        body: "<p>Test quill content</p>",
      },
      "test-post-slug"
    );
  });
});
