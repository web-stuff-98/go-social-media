import classes from "../../../styles/components/chat/VideoChat.module.scss";
import IconBtn from "../../IconBtn";
import { ImVolumeMute, ImVolumeMute2 } from "react-icons/im";
import { useState, useRef, useEffect } from "react";
import { useAuth } from "../../../context/AuthContext";
import { PeerWithID, useChat } from "./Chat";
import { useUsers } from "../../../context/UsersContext";

function PeerVideoWindow({ peerWithID }: { peerWithID: PeerWithID }) {
  const [stream, setStream] = useState<MediaStream | undefined>(undefined);

  const handleStream = (stream: MediaStream) => {
    setStream(stream);
  };

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

  const getUserName = () => getUserData(uid).username || "";

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

export default function VideoChat() {
  const { user } = useAuth();
  const { userStream, isStreaming, peers, initVideo } = useChat();

  useEffect(() => {
    initVideo();
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
