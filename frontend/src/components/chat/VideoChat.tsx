import classes from "../../styles/components/chat/VideoChat.module.scss";
import IconBtn from "../shared/IconBtn";
import { ImSpinner8, ImVolumeMute, ImVolumeMute2 } from "react-icons/im";
import { useState, useRef, useEffect, useCallback } from "react";
import { useAuth } from "../../context/AuthContext";
import { PeerWithID, useChat } from "./Chat";
import { useUsers } from "../../context/UsersContext";
import useSocket from "../../context/SocketContext";
import { IUser } from "../../interfaces/GeneralInterfaces";

export default function VideoChat({
  isRoom,
  id,
}: {
  isRoom?: boolean;
  // ID can be either another user or a room ID
  id: string;
}) {
  const { user } = useAuth();
  const { sendIfPossible } = useSocket();
  const { userStream, isStreaming, peers, leftVidChat } = useChat();
  const { getUserData } = useUsers();

  useEffect(() => {
    sendIfPossible(
      JSON.stringify({
        event_type: "VID_JOIN",
        join_id: id,
        is_room: Boolean(isRoom),
      })
    );
    return () => {
      leftVidChat(Boolean(isRoom), id);
    };
    // eslint-disable-next-line
  }, []);

  const renderUserName = (user?: IUser) => (user ? user?.username : "user");

  return (
    <div className={classes.container}>
      <span className={classes.count}>
        {peers.length} peers <br />
        {peers.map((p) => (
          <>
            {renderUserName(getUserData(p.UID))}
            <br />
          </>
        ))}
      </span>
      <div className={classes.windows}>
        {isStreaming && (
          <VideoWindow uid={user?.ID as string} stream={userStream} ownVideo />
        )}
        {peers.map((p) => (
          <PeerVideoWindow key={p.UID} peerWithID={p} />
        ))}
      </div>
    </div>
  );
}

function PeerVideoWindow({ peerWithID }: { peerWithID: PeerWithID }) {
  const [stream, setStream] = useState<MediaStream | undefined>(undefined);

  const handleStream = (stream: MediaStream) => setStream(stream);

  useEffect(() => {
    peerWithID.peer.on("stream", handleStream);
    return () => {
      peerWithID?.peer.off("stream", handleStream);
    };
    // eslint-disable-next-line
  }, []);

  return (
    <VideoWindow hide={!Boolean(stream)} uid={peerWithID.UID} stream={stream} />
  );
}

function VideoWindow({
  uid,
  stream,
  ownVideo,
  hide,
}: {
  uid: string;
  stream?: MediaStream;
  ownVideo?: boolean;
  hide?: boolean;
}) {
  const { user } = useAuth();
  const { getUserData } = useUsers();

  const [muted, setMuted] = useState(false);
  const videoRef = useRef<HTMLVideoElement | any>();

  useEffect(() => {
    if (stream) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  const getUserName = () => {
    const u = getUserData(uid);
    if (u) return u.username;
    return "";
  };

  return (
    <div
      style={hide ? { display: "none" } : {}}
      data-testid={
        ownVideo ? "Users video chat window" : `Uid ${uid}s video chat window`
      }
      className={classes.videoWindow}
    >
      <div className={classes.inner}>
        <div className={classes.topTray}>
          <div className={classes.text}>
            {ownVideo ? user?.username : getUserName()}
          </div>
          <div className={classes.icons}>
            {!ownVideo && (
              <IconBtn
                onClick={() => setMuted(!muted)}
                Icon={muted ? ImVolumeMute2 : ImVolumeMute}
                type="button"
                name={muted ? "Unmute" : "Mute"}
                ariaLabel={muted ? "Unmute" : "Mute"}
                style={{ color: "white" }}
              />
            )}
          </div>
        </div>
        <video
          style={stream ? { filter: "opacity(1)" } : {}}
          muted={ownVideo || muted}
          autoPlay
          playsInline
          ref={videoRef}
        />
        <ImSpinner8 className={classes.spinner} />
      </div>
    </div>
  );
}
