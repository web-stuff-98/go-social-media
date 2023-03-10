import { createContext, useContext, useState, useRef, useEffect } from "react";
import type { ReactNode } from "react";
import {
  instanceOfNotificationsData,
  instanceOfReceivingReturnedSignal,
  instanceOfVidAllUsers,
  instanceOfVidUserJoined,
  instanceOfVidUserLeft,
} from "../utils/DetermineSocketEvent";
import Peer from "simple-peer";
import * as process from "process";
import { useModal } from "./ModalContext";
import {
  createRoom,
  updateRoom,
  uploadRoomImage,
  deleteRoom as deleteRoomService,
} from "../services/rooms";
import useSocket from "./SocketContext";
(window as any).process = process;

export enum ChatSection {
  "MENU" = "Menu",
  "CONVS" = "Inbox",
  "ROOMS" = "Rooms",
  "ROOM" = "Room",
  "EDITOR" = "Editor",
  "MEMBERS" = "Members",
}

export const ChatContext = createContext<{
  section: ChatSection;
  setSection: (to: ChatSection) => void;

  openRoom: (id: string) => void;
  openRoomMembers: (id: string) => void;
  openRoomEditor: (id: string) => void;
  deleteRoom: (id: string) => void;

  roomId: string;
  editRoomId: string;

  userStream?: MediaStream;
  isStreaming: boolean;
  peers: PeerWithID[];
  leftVidChat: (isRoom: boolean, id: string) => void;
  toggleStream: (isRoom: boolean, id: string) => void;
  streamToggling: boolean;

  handleCreateUpdateRoom: (
    vals: { name: string; image?: File; private: boolean },
    originalImageChanged: boolean
  ) => Promise<void>;

  notifications: { type: string }[];
}>({
  section: ChatSection.MENU,
  setSection: () => {},

  openRoom: () => {},
  openRoomMembers: () => {},
  openRoomEditor: () => {},
  deleteRoom: () => {},

  roomId: "",
  editRoomId: "",

  userStream: undefined,
  isStreaming: false,
  peers: [],
  leftVidChat: () => {},
  toggleStream: () => {},
  streamToggling: false,

  handleCreateUpdateRoom: () => new Promise((r) => r()),

  notifications: [],
});

export type PeerWithID = {
  UID: string;
  peer: Peer.Instance;
};

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const { openModal } = useModal();
  const { sendIfPossible, socket } = useSocket();

  const [section, setSection] = useState<ChatSection>(ChatSection.MENU);
  const [roomId, setRoomId] = useState("");
  const [editRoomId, setEditRoomId] = useState("");
  const [notifications, setNotifications] = useState<{ type: string }[]>([]);

  const openRoomMembers = (id: string) => {
    setRoomId(id);
    setSection(ChatSection.MEMBERS);
  };

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
    vals: { name: string; image?: File; private: boolean },
    originalImageChanged: boolean
  ) => {
    if (editRoomId) {
      return new Promise<void>((resolve, reject) => {
        updateRoom({
          name: vals.name,
          private: vals.private,
          ID: editRoomId,
        })
          .then(() => {
            if (originalImageChanged && vals.image) {
              uploadRoomImage(vals.image, editRoomId)
                .then(() => resolve())
                .catch((e) => reject(e));
            } else {
              resolve();
            }
          })
          .catch((e) => reject(e));
      });
    } else {
      return new Promise<void>((resolve, reject) => {
        createRoom(vals)
          .then((id) => {
            if (vals.image) {
              uploadRoomImage(vals.image, id)
                .then(() => resolve())
                .catch((e) => reject(e));
            } else {
              resolve();
            }
          })
          .catch((e) => reject(e));
      });
    }
  };

  const handleNotifications = (notifications: { type: string }[] | null) => {
    setNotifications(notifications || []);
  };

  /////////////////////////////////////////////////////
  //////////////// VIDEO CHAT STUFF ///////////////////
  // The user is joined to the WebRTC network        //
  // automatically when they enter a conversation or //
  // a chatroom. When the user wants to stream they  //
  // click on their icon which will then just        //
  // restart the entire network for them...........  //
  //                                                 //
  // I was having extremely weird frustrating        //
  // problems which I "resolved" by just completed   //
  // restarting the peer to peer network whenever a  //
  // user enables/disables their stream. I have      //
  // spent probably 2 weeks total hours trying to    //
  // get this to behave as expected, so I will pick  //
  // the slow hacky solution instead of wasting more //
  // of my time                                      //
  /////////////////////////////////////////////////////
  const peersRef = useRef<PeerWithID[]>([]);
  const [peers, setPeers] = useState<PeerWithID[]>([]);

  const userStream = useRef<MediaStream | undefined>(undefined);
  const [isStreaming, setIsStreaming] = useState(false);

  const [streamToggling, setStreamToggling] = useState(false);

  // Instead of using addTrack or addStream (which is addTrack internally)
  // Just restart the entire connection because there are bizarre errors
  const toggleStream = async (isRoom: boolean, id: string) => {
    setStreamToggling(true);
    if (userStream.current) {
      userStream.current?.getTracks().forEach((track) => track.stop());
      userStream.current = undefined;
      setIsStreaming(false);
      leftVidChat(isRoom, id);
    } else {
      leftVidChat(isRoom, id);
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: true,
          video: true,
        });
        userStream.current = stream;
        setIsStreaming(true);
      } catch (e) {
        openModal("Message", {
          msg: `Error creating stream: ${e}`,
          err: true,
          pen: false,
        });
      }
    }
    sendIfPossible(
      JSON.stringify({
        event_type: "VID_JOIN",
        is_room: isRoom,
        join_id: id,
      })
    );
  };

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

  const createPeer = (id: string) => {
    const peer = new Peer({
      initiator: true,
      trickle: false,
      config: ICE_Config,
      ...(userStream.current ? { stream: userStream.current } : {}),
    });
    peer.on("signal", (signal) => {
      setStreamToggling(false);
      sendIfPossible(
        JSON.stringify({
          event_type: "VID_SENDING_SIGNAL_IN",
          signal_json: JSON.stringify(signal),
          user_to_signal: id,
        })
      );
    });
    return peer;
  };

  const addPeer = (incomingSignal: Peer.SignalData, callerUID: string) => {
    const peer = new Peer({
      initiator: false,
      trickle: false,
      config: ICE_Config,
      ...(userStream.current ? { stream: userStream.current } : {}),
    });
    peer.on("signal", (signal) => {
      setStreamToggling(false);
      sendIfPossible(
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
    setIsStreaming(false);
    userStream.current = undefined;
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

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  return (
    <ChatContext.Provider
      value={{
        section,
        setSection,
        roomId,
        openRoom,
        openRoomMembers,
        openRoomEditor,
        deleteRoom,
        editRoomId,
        userStream: userStream.current,
        isStreaming,
        peers,
        leftVidChat,
        toggleStream,
        streamToggling,
        handleCreateUpdateRoom,
        notifications,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
};

const useChat = () => useContext(ChatContext);
export default useChat;

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
