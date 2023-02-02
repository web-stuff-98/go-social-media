import classes from "../../styles/components/chat/Chat.module.scss";
import Inbox from "./Inbox";
import { useState, createContext, useContext, useRef, useEffect } from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import { useLocation } from "react-router-dom";
import Room from "./Room";
import Peer from "simple-peer";
import {
  instanceOfNotificationsData,
  instanceOfReceivingReturnedSignal,
  instanceOfVidAllUsers,
  instanceOfVidUserJoined,
  instanceOfVidUserLeft,
} from "../../utils/DetermineSocketEvent";
import useSocket from "../../context/SocketContext";
import { BsFillChatRightFill } from "react-icons/bs";
import {
  createRoom,
  updateRoom,
  uploadRoomImage,
  deleteRoom as deleteRoomService,
} from "../../services/rooms";
import { useModal } from "../../context/ModalContext";

import * as process from "process";
import ChatTopTray from "./ChatTopTray";
(window as any).process = process;

/*
  This contains all the video chat functions and state,
  it also contains openRoom and openRoomEditor, and the
  create/update room function. I had to put the create
  update room function in here because I couldn't get
  the test to pass otherwise.
*/

export enum ChatSection {
  "MENU" = "Menu",
  "CONVS" = "Inbox",
  "ROOMS" = "Rooms",
  "ROOM" = "Room",
  "EDITOR" = "Editor",
}

export const ChatContext = createContext<{
  section: ChatSection;
  setSection: (to: ChatSection) => void;

  openRoom: (id: string) => void;
  openRoomEditor: (id: string) => void;
  deleteRoom: (id: string) => void;

  roomId: string;
  editRoomId: string;

  initVideo: () => Promise<void>;
  userStream?: MediaStream;
  isStreaming: boolean;
  peers: PeerWithID[];
  leftVidChat: (isRoom: boolean, id: string) => void;

  handleCreateUpdateRoom: (
    vals: { name: string; image?: File },
    originalImageChanged: boolean,
    numErrs: number
  ) => void;

  notifications: { type: string }[];
}>({
  section: ChatSection.MENU,
  setSection: () => {},

  openRoom: () => {},
  openRoomEditor: () => {},
  deleteRoom: () => {},

  roomId: "",
  editRoomId: "",

  initVideo: () => new Promise((r) => r()),
  userStream: undefined,
  isStreaming: false,
  peers: [],
  leftVidChat: () => {},

  handleCreateUpdateRoom: () => {},

  notifications: [],
});

export type PeerWithID = {
  UID: string;
  peer: Peer.Instance;
};

export const useChat = () => useContext(ChatContext);

export default function Chat() {
  const { pathname } = useLocation();
  const { socket, sendIfPossible } = useSocket();
  const { openModal } = useModal();

  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);
  const [roomId, setRoomId] = useState("");
  const [editRoomId, setEditRoomId] = useState("");
  const [chatOpen, setChatOpen] = useState(false);
  const [notifications, setNotifications] = useState<{ type: string }[]>([]);

  const openRoom = (id: string) => {
    setRoomId(id);
    setSection(ChatSection.ROOM);
  };

  const openRoomEditor = (id: string) => {
    setEditRoomId(id);
    setSection(ChatSection.EDITOR);
  };

  const deleteRoom = (id: string) => {
    openModal("Confirm", {
      msg: "Are you sure you want to delete this room?",
      err: false,
      pen: false,
      confirmationCallback: async () => {
        try {
          openModal("Message", {
            msg: "Deleting room...",
            err: false,
            pen: true,
          });
          await deleteRoomService(id);
          openModal("Message", {
            msg: "Room deleted",
            err: false,
            pen: false,
          });
        } catch (e) {
          openModal("Message", {
            msg: `Error deleting room: ${e}`,
            err: true,
            pen: false,
          });
        }
      },
      cancellationCallback: () => {},
    });
  };

  const handleCreateUpdateRoom = async (
    vals: { name: string; image?: File },
    originalImageChanged: boolean,
    numErrs: number
  ) => {
    if (numErrs > 0) return;
    let id: string;
    if (editRoomId) {
      id = editRoomId;
      await updateRoom({
        name: vals.name,
        ID: editRoomId,
      });
    } else {
      id = await createRoom(vals);
    }
    if (
      (vals.image && !editRoomId) ||
      (editRoomId && originalImageChanged && vals.image)
    ) {
      await uploadRoomImage(vals.image, id);
    }
  };

  /////////////////////////////////////////////////////
  //////////////// VIDEO CHAT STUFF ///////////////////
  // The socket event handler functions were wrapped //
  // in useCallback but I got rid of that because it //
  // was breaking video chat for some reason.        //
  /////////////////////////////////////////////////////
  const userStream = useRef<MediaStream | undefined>(undefined);
  const [isStreaming, setIsStreaming] = useState(false);
  const initVideo = () =>
    new Promise<void>((resolve, reject) => {
      navigator.mediaDevices
        .getUserMedia({
          audio: true,
          video: true,
        })
        .then((stream) => {
          userStream.current = stream;
          setIsStreaming(true);
          resolve();
        })
        .catch((e) => reject(e));
    });

  const peersRef = useRef<PeerWithID[]>([]);
  const [peers, setPeers] = useState<PeerWithID[]>([]);

  const handleVidChatAllUsers = (uids: string[]) => {
    const peers: PeerWithID[] = [];
    if (uids)
      uids.forEach((id) => {
        const peer = createPeer(id);
        peersRef.current.push({
          UID: id,
          peer,
        });
        peers.push({ peer, UID: id });
      });
    setPeers(peers);
  };

  const handleVidChatUserJoined = (
    signal: Peer.SignalData,
    callerUID: string
  ) => {
    const peer = addPeer(signal, callerUID);
    setPeers((peers) => [...peers, { peer, UID: callerUID }]);
    peersRef.current.push({
      peer,
      UID: callerUID,
    });
  };

  const handleVidChatReceivingReturningSignal = (
    signal: Peer.SignalData,
    id: string
  ) => {
    const item = peersRef.current.find((p) => p.UID === id);
    setTimeout(() => {
      item?.peer.signal(signal);
    });
  };

  const handleVidChatUserLeft = (id: string) => {
    const peerRef = peersRef.current.find((p) => p.UID === id);
    peerRef?.peer.destroy();
    setPeers((peers) => peers.filter((p) => p.UID !== id));
    peersRef.current = peersRef.current.filter((p) => p.UID !== id);
  };

  const handleNotifications = (notifications: { type: string }[] | null) => {
    console.log(JSON.stringify(notifications));
    setNotifications(notifications || []);
  };

  const createPeer = (id: string) => {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      stream: userStream.current,
      config: ICE_Config,
    });
    peer.on("signal", (signal) =>
      sendIfPossible(
        JSON.stringify({
          event_type: "VID_SENDING_SIGNAL_IN",
          signal_json: JSON.stringify(signal),
          user_to_signal: id,
        })
      )
    );
    return peer;
  };

  const addPeer = (incomingSignal: Peer.SignalData, callerUID: string) => {
    const peer = new Peer({
      initiator: false,
      trickle: false,
      stream: userStream.current,
      config: ICE_Config,
    });
    peer.on("signal", (signal) =>
      sendIfPossible(
        JSON.stringify({
          event_type: "VID_RETURNING_SIGNAL_IN",
          signal_json: JSON.stringify(signal),
          caller_uid: callerUID,
        })
      )
    );
    setTimeout(() => {
      peer.signal(incomingSignal);
    });
    return peer;
  };

  const handleMessage = (e: MessageEvent) => {
    const data = JSON.parse(e.data);
    console.log(data);
    if (instanceOfReceivingReturnedSignal(data)) {
      handleVidChatReceivingReturningSignal(
        JSON.parse(data.signal_json) as Peer.SignalData,
        data.uid
      );
    }
    if (instanceOfVidAllUsers(data)) {
      handleVidChatAllUsers(data.uids);
    }
    if (instanceOfVidUserJoined(data)) {
      handleVidChatUserJoined(
        JSON.parse(data.signal_json) as Peer.SignalData,
        data.caller_uid
      );
    }
    if (instanceOfVidUserLeft(data)) {
      handleVidChatUserLeft(data.uid);
    }
    if (instanceOfNotificationsData(data)) {
      handleNotifications(JSON.parse(data.DATA));
    }
  };

  const leftVidChat = (isRoom: boolean, id: string) => {
    sendIfPossible(
      JSON.stringify({
        event_type: "VID_LEAVE",
        is_room: isRoom,
        id,
      })
    );
    peersRef.current.forEach((p) => p.peer.destroy());
    setPeers([]);
    peersRef.current = [];
    userStream.current?.getTracks().forEach((track) => track.stop());
  };

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  const notificationsContainerRef = useRef<HTMLDivElement>(null);
  return (
    <div
      style={{
        ...(pathname.includes("/blog")
          ? { bottom: "calc(var(--pagination-controls) + var(--padding))" }
          : {}),
        ...(chatOpen
          ? {}
          : { border: "none", background: "none", boxShadow: "none" }),
      }}
      className={classes.container}
    >
      {chatOpen ? (
        <>
          <ChatContext.Provider
            value={{
              setSection,
              openRoom,
              openRoomEditor,
              deleteRoom,
              initVideo,
              leftVidChat,
              section,
              roomId,
              editRoomId,
              isStreaming,
              peers,
              userStream: userStream.current,
              handleCreateUpdateRoom,
              notifications,
            }}
          >
            <ChatTopTray closeChat={() => setChatOpen(false)} />
            <div className={classes.inner}>
              {
                {
                  Inbox: <Inbox />,
                  Rooms: <Rooms />,
                  Room: <Room />,
                  Editor: <RoomEditor />,
                  Menu: <Menu />,
                }[section]
              }
            </div>
          </ChatContext.Provider>
        </>
      ) : (
        <button
          name="Open chat"
          aria-label="Open chat"
          onClick={() => setChatOpen(true)}
          className={classes.chatIconButton}
        >
          {notifications && notifications.length !== 0 && (
            <div
              ref={notificationsContainerRef}
              className={classes.notifications}
              style={{
                height: `${notificationsContainerRef.current?.clientWidth}px`,
                top: `-${
                  (notificationsContainerRef.current?.clientWidth || 0) * 0.33
                }px`,
                left: `-${
                  (notificationsContainerRef.current?.clientWidth || 0) * 0.33
                }px`,
              }}
            >
              +{notifications.length}
            </div>
          )}
          <BsFillChatRightFill />
        </button>
      )}
    </div>
  );
}

const ICE_Config = {
  iceServers: [
    {
      urls: "stun:openrelay.metered.ca:80",
    },
    {
      urls: "turn:openrelay.metered.ca:80",
      username: "openrelayproject",
      credential: "openrelayproject",
    },
  ],
};
