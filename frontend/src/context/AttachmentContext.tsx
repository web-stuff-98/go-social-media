import { useContext, createContext, useRef } from "react";
import type { ReactNode } from "react";
import { makeRequest } from "../services/makeRequest";
import { useModal } from "./ModalContext";
import getDuration from "../utils/GetVideoDuration";
import { useAuth } from "./AuthContext";

/*
    This handles uploading attachments. It used to be handle via socket but I
    couldn't get it to work because of a typo so I rewrote everything to use
    HTTP endpoints instead.
*/

const AttachmentContext = createContext<{
  // Recipient ID can be either a room or another user
  uploadAttachment: (
    file: File,
    msgId: string,
    recipientId: string,
    isRoom: boolean
  ) => void;
}>({
  uploadAttachment: () => {},
});

type FileUploadChunks = {
  ID: string;
  Chunks: Promise<ArrayBuffer>[]; //1mb chunks
};

export const AttachmentProvider = ({ children }: { children: ReactNode }) => {
  const { openModal } = useModal();
  const { user } = useAuth();

  const fileUploadChunks = useRef<FileUploadChunks[]>([]);

  //Recipient ID can be either another user or a room
  const uploadAttachment = async (
    file: File,
    msgId: string,
    recipientId: string,
    isRoom: boolean
  ) => {
    try {
      // Split attachment
      let startPointer = 0;
      let endPointer = file.size;
      while (startPointer < endPointer) {
        let newStartPointer = startPointer + 1048576;
        const i = fileUploadChunks.current.findIndex((c) => c.ID === msgId);
        if (i === -1) {
          fileUploadChunks.current = [
            ...fileUploadChunks.current,
            {
              ID: msgId,
              Chunks: [
                new Blob([
                  file.slice(startPointer, newStartPointer),
                ]).arrayBuffer(),
              ],
            },
          ];
        } else {
          fileUploadChunks.current[i] = {
            ...fileUploadChunks.current[i],
            Chunks: [
              ...fileUploadChunks.current[i].Chunks,
              new Blob([
                file.slice(startPointer, newStartPointer),
              ]).arrayBuffer(),
            ],
          };
        }
        startPointer = newStartPointer;
      }
      let length = 0;
      if (file.type === "video/mp4") {
        length = await getDuration(file);
      }
      // Send HTTP POST request to metadata endpoint
      await makeRequest(`/api/attachment/metadata/${msgId}/${recipientId}`, {
        withCredentials: true,
        data: {
          name: file.name,
          size: file.size,
          type: file.type,
          subscription_names: isRoom
            ? [`room=${recipientId}`]
            : [`inbox=${recipientId}`, `inbox=${user?.ID}`],
          length,
        },
        method: "POST",
      });
      // Upload chunks
      const c = fileUploadChunks.current.find((c) => c.ID === msgId);
      for await (const data of c?.Chunks!) {
        await makeRequest(`/api/attachment/chunk/${msgId}`, {
          withCredentials: true,
          method: "POST",
          headers: { "Content-Type": "application/octet-stream" },
          data,
        });
        await new Promise((resolve) => setTimeout(resolve, 300))
      }
    } catch (error) {
      openModal("Message", {
        msg: "Client attachment upload error: " + error,
        pen: false,
        err: true,
      });
    }
  };

  return (
    <AttachmentContext.Provider value={{ uploadAttachment }}>
      {children}
    </AttachmentContext.Provider>
  );
};

const useAttachment = () => useContext(AttachmentContext);
export default useAttachment;
