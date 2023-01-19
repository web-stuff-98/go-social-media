import {
  useState,
  useContext,
  createContext,
  useEffect,
  useCallback,
  useRef,
} from "react";
import type { ReactNode } from "react";
import { useAuth } from "./AuthContext";
import { makeRequest } from "../services/makeRequest";
import { useModal } from "./ModalContext";
import useSocket from "./SocketContext";
import { instanceOfAttachmentProgressData } from "../utils/DetermineSocketEvent";
import getDuration from "../utils/GetVideoDuration";

/*
    This handles uploading attachments. It used to be handle via socket but I
    couldn't get it to work so I'm uploading chunks to a HTTP endpoint now instead.
*/

const AttachmentContext = createContext<{
  uploadAttachment: (file: File, msgId: string, recipientId: string) => void;
}>({
  uploadAttachment: () => {},
});

type RequestNextMbEvent = {
  ID: string;
  index: number;
};

type FileUploadChunks = {
  ID: string;
  Chunks: Promise<ArrayBuffer>[][]; //1mb chunks
};

export const AttachmentProvider = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  const { socket } = useSocket();
  const { openModal } = useModal();

  const [fileSocket, setFileSocket] = useState<WebSocket>();
  const [failed, setFailed] = useState<string[]>([]);

  const fileUploadChunks = useRef<FileUploadChunks[]>([]);

  //Recipient ID can be either another user or a room
  const uploadAttachment = async (
    file: File,
    msgId: string,
    recipientId: string
  ) => {
    try {
      // Split attachment
      let startPointer = 0;
      let endPointer = file.size;
      let promises: Promise<ArrayBuffer>[] = [];
      while (startPointer < endPointer) {
        let newStartPointer = startPointer + 8072;
        promises.push(
          new Blob([
            msgId,
            file.slice(startPointer, newStartPointer),
          ]).arrayBuffer()
        );
        if (
          promises.length >= 129 ||
          newStartPointer >= file.size ||
          file.size <= 1051672
        ) {
          const i = fileUploadChunks.current.findIndex((c) => c.ID === msgId);
          if (i === -1) {
            fileUploadChunks.current = [
              ...fileUploadChunks.current,
              {
                ID: msgId,
                Chunks: [promises],
              },
            ];
          } else {
            fileUploadChunks.current[i] = {
              ...fileUploadChunks.current[i],
              Chunks: [...fileUploadChunks.current[i].Chunks, promises],
            };
          }
          promises = [];
        }
        startPointer = newStartPointer;
      }
      let length = 0;
      if (file.type === "video/mp4") {
        length = await getDuration(file);
      }
      // Send HTTP POST request to metadata endpoint
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
    } catch (error) {
      openModal("Message", {
        msg: "Client attachment upload error: " + error,
        pen: false,
        err: true,
      });
    }
  };

  const uploadChunk = (msgId: string, index: number) => {
    const c = fileUploadChunks.current.find((c) => c.ID === msgId);
    console.log("Chunks index:", index);
    console.log("Chunks length:", c?.Chunks.length);
    console.log("Length:", c?.Chunks[index]!.length);
    for (const buff of c?.Chunks[index]!) {
      buff.then((b) => {
        fileSocket?.send(b);
      });
    }
    // Done sending 1mb chunk. If there are more 1mb chunks then
    // wait for the response from the server before sending the
    // next, otherwise send the finish signal (msgId on its own)
    if (c?.Chunks.length === 1) {
      fileSocket?.send(msgId);
    } else if (index === c?.Chunks.length! - 1) {
      fileSocket?.send(msgId);
    }
  };

  return (
    <AttachmentContext.Provider
      value={{ uploadAttachment }}
    >
      {children}
    </AttachmentContext.Provider>
  );
};

const useAttachment = () => useContext(AttachmentContext);
export default useAttachment;
