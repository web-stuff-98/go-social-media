import { screen, render } from "@testing-library/react";
import { unmountComponentAtNode } from "react-dom";
import { act } from "react-dom/test-utils";
import { AuthContext } from "../../context/AuthContext";
import { SocketContext } from "../../context/SocketContext";
import { ChatContext } from "./Chat";
import VideoChat from "./VideoChat";

let container = null;

const createMockPeer = () => ({
  off: jest.fn(),
  on: jest.fn(),
});

const createMockPeers = () => {
  let peers = [];
  let i = 3;
  while (i < 3) {
    peers.push({ uid: `${i + 1}`, peer: createMockPeer() });
    i++;
  }
  return peers;
};

const mockPeers = createMockPeers();

const mockUser = {
  ID: "1",
  username: "Test user",
};

beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
});

afterEach(() => {
  unmountComponentAtNode(container);
  container.remove();
  container = null;
});

describe("video chat windows", () => {
  test("initVideo should have been invoked. sendIfPossible should have been invoked with the correct parameters. The video chat windows for the peer and the current user should be present", async () => {
    const mockSendIfPossible = jest.fn();
    const mockInitVideo = jest.fn().mockReturnValue(new Promise((r) => r()));

    await act(async () => {
      render(
        <SocketContext.Provider value={{ sendIfPossible: mockSendIfPossible }}>
          <ChatContext.Provider
            value={{
              initVideo: mockInitVideo,
              leftVidChat: jest.fn(),
              peers: mockPeers,
              isStreaming: true,
            }}
          >
            <AuthContext.Provider value={{ user: mockUser }}>
              <VideoChat id="1" />
            </AuthContext.Provider>
          </ChatContext.Provider>
        </SocketContext.Provider>,
        container
      );
    });

    expect(mockInitVideo).toHaveBeenCalled();
    expect(mockSendIfPossible).toHaveBeenCalledWith(
      JSON.stringify({
        event_type: "VID_JOIN",
        join_id: "1",
        is_room: false,
      })
    );

    expect(screen.getByTestId("Users video chat window")).toBeInTheDocument();

    for await (const peer of mockPeers) {
      const peerVideoWindow = await screen.findByTestId(
        `Uid ${peer.uid}s video chat window`
      );
      expect(peerVideoWindow).toBeInTheDocument();
    }
  });
});
