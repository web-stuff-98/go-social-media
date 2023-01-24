import classes from "../../styles/components/chat/Chat.module.scss";
import IconBtn from "../shared/IconBtn";
import { IoMdClose } from "react-icons/io";
import Conversations from "./Conversations";
import { MdMenu } from "react-icons/md";
import { useState, createContext, useContext, useRef, useEffect } from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import { useLocation } from "react-router-dom";
import Room from "./Room";
import { useModal } from "../../context/ModalContext";
import Peer from "simple-peer";
import {
  instanceOfReceivingReturnedSignal,
  instanceOfVidAllUsers,
  instanceOfVidUserJoined,
  instanceOfVidUserLeft,
} from "../../utils/DetermineSocketEvent";
import useSocket from "../../context/SocketContext";
import { BsFillChatRightFill } from "react-icons/bs";

import * as process from "process";
(window as any).process = process;

export enum ChatSection {
  "MENU" = "Menu",
  "CONVS" = "Conversations",
  "ROOMS" = "Rooms",
  "ROOM" = "Room",
  "EDITOR" = "Editor",
}

const ChatContext = createContext<{
  section: ChatSection;
  setSection: (to: ChatSection) => void;

  openRoom: (id: string) => void;
  openRoomEditor: (id: string) => void;

  roomId: string;
  editRoomId: string;

  initVideo: (cb: Function) => void;
  userStream?: MediaStream;
  isStreaming: boolean;
  peers: PeerWithID[];
  leftVidChat: (isRoom: boolean, id: string) => void;
}>({
  section: ChatSection.MENU,
  setSection: () => {},

  openRoom: () => {},
  openRoomEditor: () => {},

  roomId: "",
  editRoomId: "",

  initVideo: async () => {},
  userStream: undefined,
  isStreaming: false,
  peers: [],
  leftVidChat: () => {},
});

export type PeerWithID = {
  UID: string;
  peer: Peer.Instance;
};

export const useChat = () => useContext(ChatContext);

export default function Chat() {
  const { pathname } = useLocation();
  const { openModal } = useModal();
  const { socket } = useSocket();

  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);
  const [roomId, setRoomId] = useState("");
  const [editRoomId, setEditRoomId] = useState("");
  const [chatOpen, setChatOpen] = useState(false);

  const openRoom = (id: string) => {
    setRoomId(id);
    setSection(ChatSection.ROOM);
  };

  const openRoomEditor = (id: string) => {
    setEditRoomId(id);
    setSection(ChatSection.EDITOR);
  };

  //////////////// VIDEO CHAT STUFF \\\\\\\\\\\\\\\\
  const userStream = useRef<MediaStream | undefined>(undefined);
  const [isStreaming, setIsStreaming] = useState(false);
  const initVideo = async (cb: Function) => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: true,
        video: true,
      });
      userStream.current = stream;
      setIsStreaming(true);
      cb();
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
    }
  };

  const peersRef = useRef<PeerWithID[]>([]);
  const [peers, setPeers] = useState<PeerWithID[]>([]);

  const handleVidChatAllUsers = (uids: string[]) => {
    const peers: PeerWithID[] = [];
    console.log("All users");
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
    console.log("User joined");
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
    console.log("Receiving returned signal");
    const item = peersRef.current.find((p) => p.UID === id);
    setTimeout(() => {
      item?.peer.signal(signal);
    });
  };

  const handleVidChatUserLeft = (id: string) => {
    console.log("User left");
    const peerRef = peersRef.current.find((p) => p.UID === id);
    peerRef?.peer.destroy();
    setPeers((peers) => peers.filter((p) => p.UID !== id));
    peersRef.current = peersRef.current.filter((p) => p.UID !== id);
  };

  const createPeer = (id: string) => {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      stream: userStream.current,
      config: ICE_Config,
    });
    peer.on("signal", (signal) =>
      socket?.send(
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
    console.log("Adding peer");
    peer.on("signal", (signal) =>
      socket?.send(
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
  };

  const leftVidChat = (isRoom: boolean, id: string) => {
    socket?.send(
      JSON.stringify({
        event_type: "VID_LEAVE",
        is_room: isRoom,
        id,
      })
    );
    peersRef.current.forEach((p) => p.peer.destroy());
    setPeers([]);
    peersRef.current = [];
  };

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

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
          <div className={classes.topTray}>
            {section}
            <div className={classes.icons}>
              {section !== ChatSection.MENU && (
                <IconBtn
                  onClick={() => setSection(ChatSection.MENU)}
                  name="Chat menu"
                  ariaLabel="Chat menu"
                  Icon={MdMenu}
                />
              )}
              <IconBtn
                name="Close chat"
                ariaLabel="Close chat"
                onClick={() => setChatOpen(false)}
                Icon={IoMdClose}
              />
            </div>
          </div>
          <ChatContext.Provider
            value={{
              section,
              setSection,
              roomId,
              editRoomId,
              openRoom,
              openRoomEditor,
              initVideo,
              isStreaming,
              peers,
              leftVidChat,
              userStream: userStream.current,
            }}
          >
            <div className={classes.inner}>
              {
                {
                  Conversations: <Conversations />,
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
          aria-label="Open chat"
          onClick={() => setChatOpen(true)}
          className={classes.chatIconButton}
        >
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
