import {
  useState,
  useContext,
  createContext,
  useEffect,
  useCallback,
} from "react";
import type { ReactNode } from "react";
import { useAuth } from "./AuthContext";
import { makeRequest } from "../services/makeRequest";
import { useModal } from "./ModalContext";
import useSocket from "./SocketContext";
import { instanceOfAttachmentProgressData } from "../utils/DetermineSocketEvent";
import getDuration from "../utils/GetVideoDuration";

/*
    This handles the websocket connection to the file socket endpoint. The connection is
    only opened when the user is logged in.
*/

const FileSocketContext = createContext<{
  fileSocket?: WebSocket;
  connectFileSocket: () => void;
  // RecipientID can be either user or room
  uploadAttachment: (file: File, msgId: string, recipientId: string) => void;
}>({
  fileSocket: undefined,
  connectFileSocket: () => {},
  uploadAttachment: () => {},
});

export const FileSocketProvider = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  const { socket } = useSocket();
  const { openModal } = useModal();

  const [fileSocket, setFileSocket] = useState<WebSocket>();
  const [failed, setFailed] = useState<string[]>([]);

  const connectFileSocket = () => {
    const fileSocket = new WebSocket(
      process.env.NODE_ENV === "development"
        ? "ws://localhost:8080/api/file/ws"
        : "wss://go-social-media-js.herokuapp.com/api/file/ws"
    );
    fileSocket.binaryType = "arraybuffer";
    setFileSocket(fileSocket);
  };

  //Recipient ID can be either another user or a room
  const uploadAttachment = async (
    file: File,
    msgId: string,
    recipientId: string
  ) => {
    try {
      let isVideo = false;
      let length = 0;
      if (file.type === "video/mp4") {
        isVideo = true;
        length = await getDuration(file);
      }
      // First send HTTP POST request to metadata endpoint
      await makeRequest(`/api/attachment/${msgId}/${recipientId}`, {
        withCredentials: true,
        data: {
          name: file.name,
          size: file.size,
          type: file.type,
          length,
        },
        method: "POST",
      });
      // Upload attachment in chunks (with the first 24 bytes being the msg id)
      let startPointer = 0;
      let endPointer = file.size;
      let promises = [];
      while (startPointer < endPointer) {
        let newStartPointer = startPointer + 262120;
        promises.push(
          new Blob([
            msgId,
            file.slice(startPointer, newStartPointer),
          ]).arrayBuffer()
        );
        startPointer = newStartPointer;
      }
      for await (const buff of promises) {
        if (failed.includes(msgId)) {
          setFailed((f) => [...f.filter((f) => f !== msgId)]);
          return;
        }
        await new Promise<void>((resolve) => {
          fileSocket?.send(buff);
          setTimeout(() => {
            resolve();
          }, 500);
        });
      }
      await new Promise<void>((r) => setTimeout(() => r(), 100));
      // When the attachment is finished uploading, send the message ID on its own, that way the server knows its done
      fileSocket?.send(msgId);
    } catch (error) {
      openModal("Message", {
        msg: "Client attachment upload error: " + error,
        pen: false,
        err: true,
      });
    }
  };

  useEffect(() => {
    if (user) connectFileSocket();
    else setFileSocket(undefined);
    return () => {
      if (fileSocket) setFileSocket(undefined);
    };
  }, [user]);

  // Watch for failures, if an attachment failed stop sending bytes
  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfAttachmentProgressData(data)) {
      if (data.DATA.ID && data.DATA.failed) {
        setFailed((f) => [...f, data.DATA.ID]);
      }
    }
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
  }, [socket]);

  return (
    <FileSocketContext.Provider
      value={{ fileSocket, connectFileSocket, uploadAttachment }}
    >
      {children}
    </FileSocketContext.Provider>
  );
};

const useFileSocket = () => useContext(FileSocketContext);
export default useFileSocket;
