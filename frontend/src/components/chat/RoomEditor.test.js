import { screen, render, fireEvent, waitFor } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { ChatContext } from "./Chat";
import RoomEditor from "./RoomEditor";
import * as roomServices from "../../services/rooms";
import { AuthContext } from "../../context/AuthContext";
import { act } from "react-dom/test-utils";

roomServices.createRoom = jest.fn().mockResolvedValueOnce("1");
roomServices.getRoom = jest.fn().mockReturnValue({
  ID: "1",
  name: "Room name",
  author_id: "1",
  img_blur: "placeholder",
  img_url: "",
});
roomServices.getRandomRoomImage = jest.fn().mockResolvedValueOnce(
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

const mockUser = {
  ID: "1",
  username: "Test user",
};

describe("room editor", () => {
  test("should render with all the inputs and buttons present", async () => {
    render(
      <ChatContext.Provider value={{ editRoomId: undefined }}>
        <RoomEditor />
      </ChatContext.Provider>,
      container
    );

    const roomNameInput = screen.getByRole("textbox", {
      name: "Room name",
      hidden: true,
    });
    const roomImageBtn = screen.getByTestId("Image file button");
    const randomRoomImageBtn = screen.getByRole("button", {
      name: "Random image",
      hidden: true,
    });
    const submitBtn = screen.getByRole("button", {
      name: "Create room",
      hidden: true,
    });
    const imgPreviewElem = screen.getByTestId("Image preview");

    expect(roomNameInput).toBeInTheDocument();
    expect(roomImageBtn).toBeInTheDocument();
    expect(randomRoomImageBtn).toBeInTheDocument();
    expect(submitBtn).toBeInTheDocument();
    expect(imgPreviewElem).toBeInTheDocument();
  });

  test("filling out the inputs, pressing random image button and clicking submit should invoke the handleCreateUpdateRoom function. The image preview should also be present", async () => {
    const mockHandleCreateUpdateRoom = jest.fn();

    await act(async () => {
      render(
        <AuthContext.Provider value={{ user: mockUser }}>
          <ChatContext.Provider
            value={{
              editRoomId: undefined,
              handleCreateUpdateRoom: mockHandleCreateUpdateRoom,
            }}
          >
            <RoomEditor />
          </ChatContext.Provider>
        </AuthContext.Provider>,
        container
      );
    });

    const roomNameInput = screen.getByRole("textbox", {
      name: "Room name",
      hidden: true,
    });
    const randomRoomImageBtn = screen.getByRole("button", {
      name: "Random image",
      hidden: true,
    });
    const submitBtn = screen.getByRole("button", {
      name: "Create room",
      hidden: true,
    });

    await act(async () => {
      fireEvent.change(roomNameInput, { target: { value: "Test room" } });
    });

    await act(async () => {
      randomRoomImageBtn.click();
    });

    expect(roomServices.getRandomRoomImage).toHaveBeenCalled();
    const imgPreviewElem = screen.getByTestId("Image preview");
    expect(imgPreviewElem).toBeInTheDocument();

    await act(async () => {
      submitBtn.click();
    });

    expect(mockHandleCreateUpdateRoom).toHaveBeenCalled();
  });

  test("when editRoomId is present, filling out the inputs, pressing random image button and clicking submit should invoke the handleCreateUpdateRoom function. The image preview should also be present", async () => {
    const mockHandleCreateUpdateRoom = jest.fn();

    await act(async () => {
      render(
        <AuthContext.Provider value={{ user: mockUser }}>
          <ChatContext.Provider
            value={{
              editRoomId: "1",
              handleCreateUpdateRoom: mockHandleCreateUpdateRoom,
            }}
          >
            <RoomEditor />
          </ChatContext.Provider>
        </AuthContext.Provider>,
        container
      );
    });

    expect(roomServices.getRoom).toHaveBeenCalled();

    // jest couldn't find this with getByRole for some reason......?
    const roomNameInput = screen.getByTestId("name");
    const randomRoomImageBtn = screen.getByRole("button", {
      name: "Random image",
      hidden: true,
    });
    const submitBtn = screen.getByRole("button", {
      name: "Update room",
      hidden: true,
    });

    await act(async () => {
      fireEvent.change(roomNameInput, { target: { value: "Test room" } });
    });

    await act(async () => {
      randomRoomImageBtn.click();
    });

    expect(roomServices.getRandomRoomImage).toHaveBeenCalled();
    const imgPreviewElem = screen.getByTestId("Image preview");
    expect(imgPreviewElem).toBeInTheDocument();

    await act(async () => {
      submitBtn.click();
    });

    expect(mockHandleCreateUpdateRoom).toHaveBeenCalled();
  });
});
