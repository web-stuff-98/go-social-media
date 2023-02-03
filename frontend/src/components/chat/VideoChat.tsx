import classes from "../../styles/components/chat/VideoChat.module.scss";
import IconBtn from "../shared/IconBtn";
import { ImSpinner8, ImVolumeMute, ImVolumeMute2 } from "react-icons/im";
import { useState, useRef, useEffect } from "react";
import { useAuth } from "../../context/AuthContext";
import { PeerWithID, useChat } from "./Chat";
import { useUsers } from "../../context/UsersContext";
import useSocket from "../../context/SocketContext";
import { useModal } from "../../context/ModalContext";

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
  const { openModal } = useModal();
  const { userStream, isStreaming, peers, initVideo, leftVidChat } = useChat();

  const setupVideo = async () => {
    try {
      await initVideo();
      sendIfPossible(
        JSON.stringify({
          event_type: "VID_JOIN",
          join_id: id,
          is_room: Boolean(isRoom),
        })
      );
    } catch (e) {
      openModal("Message", {
        msg: `Failed to open video chat: ${e}`,
        err: true,
        pen: false,
      });
    }
  };

  useEffect(() => {
    setupVideo();
    return () => {
      leftVidChat(Boolean(isRoom), id);
    };
    // eslint-disable-next-line
  }, []);

  return (
    <div className={classes.container}>
      <span className={classes.count}>{peers.length} peers</span>
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
    <div
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
