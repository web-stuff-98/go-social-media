import { useState, useContext, createContext, useEffect } from "react";
import type { ReactNode } from "react";
import { IResMsg } from "../components/ResMsg";
import { useAuth } from "./AuthContext";

/*
    This handles the websocket connection to the file socket endpoint. The connection is
    only opened when the user is logged in.
*/

const FileSocketContext = createContext<{
  fileSocket?: WebSocket;
  connectFileSocket: () => void;
  uploadAttachment: (file: File, msgId: string) => void;
}>({
  fileSocket: undefined,
  connectFileSocket: () => {},
  uploadAttachment: () => {},
});

export const FileSocketProvider = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();

  const [fileSocket, setFileSocket] = useState<WebSocket>();
  const [fileUploadsStatus, setFileUploadsStatus] = useState<IResMsg[]>([]);

  const connectFileSocket = () => {
    const fileSocket = new WebSocket(
      process.env.NODE_ENV === "development"
        ? "ws://localhost:8080/api/file/ws"
        : "wss://go-social-media-js.herokuapp.com/api/file/ws"
    );
    fileSocket.binaryType = "arraybuffer";
    setFileSocket(fileSocket);
  };

  // Upload attachment in 4kb chunks (with the first 24 bytes being the msg id)
  const uploadAttachment = async (file: File, msgId: string) => {
    let startPointer = 0;
    let endPointer = file.size;
    let promises = [];
    while (startPointer < endPointer) {
      let newStartPointer = startPointer + 4072;
      promises.push(
        new Blob([
          msgId,
          file.slice(startPointer, newStartPointer),
        ]).arrayBuffer()
      );
      startPointer = newStartPointer;
    }
    for await (const buff of promises) {
      fileSocket?.send(buff);
    }
    // When the attachment is finished uploading, send the message ID on its own, that way the server knows its done
    fileSocket?.send(msgId);
    window.alert("Sent all bytes");
  };

  useEffect(() => {
    if (user) connectFileSocket();
    else setFileSocket(undefined);
    return () => {
      if (fileSocket) setFileSocket(undefined);
    };
  }, [user]);

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
