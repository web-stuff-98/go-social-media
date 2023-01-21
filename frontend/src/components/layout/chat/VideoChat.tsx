import classes from "../../../styles/components/chat/VideoChat.module.scss";
import IconBtn from "../../IconBtn";
import { ImVolumeMute, ImVolumeMute2 } from "react-icons/im";
import { useState, useRef, useEffect } from "react";
import { useAuth } from "../../../context/AuthContext";
import { PeerWithID, useChat } from "./Chat";
import { useUsers } from "../../../context/UsersContext";
import useSocket from "../../../context/SocketContext";

export default function VideoChat({
  isRoom,
  id,
}: {
  isRoom: boolean;
  // ID can be either another user or a room ID
  id: string;
}) {
  const { user } = useAuth();
  const { socket } = useSocket();
  const { userStream, isStreaming, peers, initVideo } = useChat();

  useEffect(() => {
    initVideo();
    socket?.send(
      JSON.stringify({
        event_type: "VID_JOIN",
        join_id: id,
        is_room: isRoom,
      })
    );
    return () => {
      socket?.send(
        JSON.stringify({
          event_type: "VID_LEAVE",
          is_room: isRoom
        })
      );
    };
  }, []);

  return (
    <div className={classes.container}>
      {isStreaming && (
        <VideoWindow uid={user?.ID as string} stream={userStream} ownVideo />
      )}
      {peers.map((p) => (
        <PeerVideoWindow peerWithID={p} />
      ))}
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
  }, []);

  return <VideoWindow uid={peerWithID.UID} stream={stream} />;
}

function VideoWindow({
  uid,
  stream,
  ownVideo,
}: {
  uid: string;
  stream?: MediaStream;
  ownVideo?: boolean;
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
    <div className={classes.videoWindow}>
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
        <video muted={ownVideo || muted} autoPlay playsInline ref={videoRef} />
      </div>
    </div>
  );
}
