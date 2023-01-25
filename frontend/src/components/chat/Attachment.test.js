import { screen, render } from "@testing-library/react";
import Attachment from "./Attachment";

describe("the attachment component", () => {
  test("should render a progress bar while the upload is pending", () => {
    render(
      <Attachment progressData={{ ratio: 0.5, failed: false, pending: true }} />
    );

    const progressElement = screen.getByTestId("Progress bar");

    expect(progressElement.toBeInTheDocument);
  });

  test("the attachment is complete and if the attachment is not recognized as an image a button with a hidden download link should be present. Clicking on the button should trigger a fetch request.", () => {
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

    expect(buttonElement).toBeInTheDocument;
    expect(linkElement).toBeInTheDocument;

    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    buttonElement.click();

    expect(axiosSpy).toHaveBeenCalled;

    global.fetch.mockClear();
  });

  test("the attachment is complete and is an image, the image should be present, and a fetch request should also be made to retrieve the image binary", () => {
    render(
      <Attachment
        metaData={{ type: "image/jpg" }}
        progressData={{ ratio: 1, failed: false, pending: false }}
      />
    );

    const imgElement = screen.getByRole("img");

    expect(imgElement).toBeInTheDocument;

    const axiosSpy = jest
      .spyOn(global, "fetch")
      .mockImplementation(
        async () => await new Promise((resolve) => setTimeout(resolve, 100))
      );

    expect(axiosSpy).toHaveBeenCalled;

    global.fetch.mockClear();
  });

  test("should render a message saying the attachment failed to upload", () => {
    render(
      <Attachment progressData={{ ratio: 0, failed: true, pending: false }} />
    );

    const containerElem = screen.getByText("Upload failed");

    expect(containerElem).toBeInTheDocument;
  });
});
