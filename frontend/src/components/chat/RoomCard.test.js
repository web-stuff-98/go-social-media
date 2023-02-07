import { screen, render, waitFor } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { ChatContext } from "./Chat";
import RoomCard from "./RoomCard";
import * as roomServices from "../../services/rooms";
import { act } from "react-dom/test-utils";
import { SocketContext } from "../../context/SocketContext";
import { AuthContext } from "../../context/AuthContext";

const b64 =
  "data:image/jpeg;base64,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";

roomServices.getRoomImage = jest.fn().mockResolvedValueOnce(b64);

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

const mockRoomCard = {
  ID: "1",
  name: "Test room",
  author_id: "1",
  img_blur: "placeholder",
  img_url: "placeholder url",
  can_access: true,
};

const mockUser = {
  ID: "1",
  username: "Test user",
};

describe("rooms menu card", () => {
  test("should get the image with getRoomImage, render the room card with the name, icons and blur placeholder, openSubscription should have been called", async () => {
    const mockOpenSubscription = jest.fn();
    const mockCloseSubscription = jest.fn();

    await act(async () => {
      render(
        <SocketContext.Provider
          value={{
            openSubscription: mockOpenSubscription,
            closeSubscription: mockCloseSubscription,
          }}
        >
          <AuthContext.Provider value={{ user: mockUser }}>
            <ChatContext.Provider
              value={{ openRoom: jest.fn(), openRoomEditor: jest.fn() }}
            >
              <RoomCard r={mockRoomCard} />
            </ChatContext.Provider>
          </AuthContext.Provider>
        </SocketContext.Provider>,
        container
      );
    });

    expect(roomServices.getRoomImage).toHaveBeenCalled();
    expect(mockOpenSubscription).toHaveBeenCalled();

    expect(screen.getByTestId("Container")).toHaveStyle(
      "background-image: url(placeholder)"
    );

    expect(screen.getByText(mockRoomCard.name)).toBeInTheDocument();

    const editRoomBtn = screen.getByRole("button", {
      name: "Edit room",
      hidden: true,
    });

    const enterRoomBtn = screen.getByRole("button", {
      name: "Enter room",
      hidden: true,
    });

    expect(editRoomBtn).toBeInTheDocument();
    expect(enterRoomBtn).toBeInTheDocument();
  });

  test("clicking on the edit room button should invoke the openRoomEditor function", async () => {
    const mockOpenSubscription = jest.fn();
    const mockCloseSubscription = jest.fn();
    const mockOpenRoomEditor = jest.fn();

    await act(async () => {
      render(
        <SocketContext.Provider
          value={{
            openSubscription: mockOpenSubscription,
            closeSubscription: mockCloseSubscription,
          }}
        >
          <AuthContext.Provider value={{ user: mockUser }}>
            <ChatContext.Provider
              value={{
                openRoom: jest.fn(),
                openRoomEditor: mockOpenRoomEditor,
              }}
            >
              <RoomCard r={mockRoomCard} />
            </ChatContext.Provider>
          </AuthContext.Provider>
        </SocketContext.Provider>,
        container
      );
    });

    const editRoomBtn = screen.getByRole("button", {
      name: "Edit room",
      hidden: true,
    });

    expect(editRoomBtn).toBeInTheDocument();
    editRoomBtn.click();
    expect(mockOpenRoomEditor).toHaveBeenCalledWith(mockRoomCard.ID);
  });

  test("clicking on the enter room button should invoke the openRoom function", async () => {
    const mockOpenSubscription = jest.fn();
    const mockCloseSubscription = jest.fn();
    const mockOpenRoom = jest.fn();

    await act(async () => {
      render(
        <SocketContext.Provider
          value={{
            openSubscription: mockOpenSubscription,
            closeSubscription: mockCloseSubscription,
          }}
        >
          <AuthContext.Provider value={{ user: mockUser }}>
            <ChatContext.Provider
              value={{ openRoom: mockOpenRoom, openRoomEditor: jest.fn() }}
            >
              <RoomCard r={mockRoomCard} />
            </ChatContext.Provider>
          </AuthContext.Provider>
        </SocketContext.Provider>,
        container
      );
    });

    const enterRoomBtn = screen.getByRole("button", {
      name: "Enter room",
      hidden: true,
    });

    expect(enterRoomBtn).toBeInTheDocument();
    enterRoomBtn.click();
    expect(mockOpenRoom).toHaveBeenCalledWith(mockRoomCard.ID);
  });
});
