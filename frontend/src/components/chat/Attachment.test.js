import { screen, render } from "@testing-library/react";
import Attachment from "./Attachment";

jest.mock("axios");

describe("the attachment component", () => {
  test("should render a progress bar while the upload is pending", () => {
    render(
      <Attachment progressData={{ ratio: 0.5, failed: false, pending: true }} />
    );

    const progressElement = screen.getByTestId("Progress bar");

    expect(progressElement.toBeInTheDocument);
  });

  test("the attachment is complete and if the attachment is not recognized as an image a button with a hidden download link should be present", () => {
    render(
      <Attachment
        metaData={{ type: "application/pdf" }}
        progressData={{ ratio: 1, failed: false, pending: false }}
      />
    );

    const buttonElement = screen.getByRole("button", {
      name: "Download attachment",
    });
    const linkElement = screen.getByRole("link", { hidden: true });

    expect(buttonElement).toBeInTheDocument();
    expect(linkElement).toBeInTheDocument();
  });

  test("the attachment is complete and is an image, the image should be present", () => {
    render(
      <Attachment
        metaData={{ type: "image/jpg" }}
        progressData={{ ratio: 1, failed: false, pending: false }}
      />
    );

    const imgElement = screen.getByRole("img");

    expect(imgElement).toBeInTheDocument();
  });

  test("should render a message saying the attachment failed to upload", () => {
    render(
      <Attachment progressData={{ ratio: 0, failed: true, pending: false }} />
    );

    const containerElem = screen.getByText("Upload failed");

    expect(containerElem).toBeInTheDocument();
  });
});
