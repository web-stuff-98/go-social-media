import classes from "../../../styles/components/chat/Chat.module.scss";
import IconBtn from "../../IconBtn";
import { IoMdClose } from "react-icons/io";
import Conversations from "./Conversations";
import { MdMenu } from "react-icons/md";
import {
  useState,
  createContext,
  useContext,
  useRef,
  useCallback,
  useEffect,
} from "react";
import Menu from "./Menu";
import RoomEditor from "./RoomEditor";
import Rooms from "./Rooms";
import { useLocation } from "react-router-dom";
import Room from "./Room";
import { useModal } from "../../../context/ModalContext";
import Peer from "simple-peer";
import {
  instanceOfReceivingReturnedSignal,
  instanceOfVidAllUsers,
  instanceOfVidUserJoined,
  instanceOfVidUserLeft,
} from "../../../utils/DetermineSocketEvent";
import useSocket from "../../../context/SocketContext";

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

  initVideo: () => void;
  userStream?: MediaStream;
  isStreaming: boolean;
  peers: PeerWithID[];
}>({
  section: ChatSection.MENU,
  setSection: () => {},

  openRoom: () => {},
  openRoomEditor: () => {},

  roomId: "",
  editRoomId: "",

  initVideo: () => {},
  userStream: undefined,
  isStreaming: false,
  peers: [],
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
  const initVideo = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: true,
        video: true,
      });
      userStream.current = stream;
      setIsStreaming(true);
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

  const handleVidChatAllUsers = useCallback((uids: string[]) => {
    const peers: PeerWithID[] = [];
    uids.forEach((id) => {
      const peer = createPeer(id);
      peersRef.current.push({
        UID: id,
        peer,
      });
      peers.push({ peer, UID: id });
    });
    setPeers(peers);
  }, []);

  const handleVidChatUserJoined = useCallback(
    (signal: Peer.SignalData, callerUID: string) => {
      const peer = addPeer(signal, callerUID);
      setPeers((peers) => [...peers, { peer, UID: callerUID }]);
      peersRef.current.push({
        peer,
        UID: callerUID,
      });
    },
    []
  );

  const handleVidChatReceivingReturningSignal = useCallback(
    (signal: Peer.SignalData, id: string) => {
      const item = peersRef.current.find((p) => p.UID === id);
      setTimeout(() => {
        item?.peer.signal(signal);
      });
    },
    []
  );

  const handleVidChatUserLeft = useCallback((id: string) => {
    const peerRef = peersRef.current.find((p) => p.UID === id);
    peerRef?.peer.destroy();
    setPeers((peers) => peers.filter((p) => p.UID !== id));
    peersRef.current = peersRef.current.filter((p) => p.UID !== id);
  }, []);

  const createPeer = (id: string) => {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      stream: userStream.current,
      config: ICE_Config,
    });
    peer.on("signal", (signal) => {
      socket?.send(
        JSON.stringify({
          event_type: "VID_SENDING_SIGNAL_IN",
          signal_json: JSON.stringify(signal),
          user_to_signal: id,
        })
      );
    });
    return peer;
  };

  const addPeer = useCallback(
    (incomingSignal: Peer.SignalData, callerUID: string) => {
      const peer = new Peer({
        initiator: false,
        trickle: false,
        stream: userStream.current,
        config: ICE_Config,
      });
      peer.on("signal", (signal) => {
        socket?.send(
          JSON.stringify({
            event_type: "VID_RETURNING_SIGNAL_IN",
            signal_json: JSON.stringify(signal),
            caller_uid: callerUID,
          })
        );
      });
      setTimeout(() => {
        peer.signal(incomingSignal);
      });
      return peer;
    },
    []
  );

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (instanceOfReceivingReturnedSignal(data)) {
      handleVidChatReceivingReturningSignal(
        JSON.parse(data.signal_json),
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
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <div
      style={
        pathname.includes("/blog")
          ? { bottom: "calc(var(--pagination-controls) + var(--padding))" }
          : {}
      }
      className={classes.container}
    >
      <div className={classes.topTray}>
        {section}
        <div className={classes.icons}>
          <IconBtn
            onClick={() => setSection(ChatSection.MENU)}
            name="Chat menu"
            ariaLabel="Chat menu"
            Icon={MdMenu}
          />
          <IconBtn name="Close chat" ariaLabel="Close chat" Icon={IoMdClose} />
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
