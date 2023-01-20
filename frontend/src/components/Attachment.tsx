import classes from "../styles/components/Attachment.module.scss";
import ProgressBar from "./ProgressBar";
import { useMemo, useRef } from "react";
import { BiError } from "react-icons/bi";
import { AiOutlineDownload } from "react-icons/ai";
import { baseURL } from "../services/makeRequest";

export interface IMsgAttachmentProgress {
  ratio: number;
  failed: boolean;
  pending: boolean;
}

export interface IAttachmentData {
  name: string;
  type: string;
  size: number;
  length: number;
}

export default function Attachment({
  progressData: { failed, ratio, pending },
  metaData,
  reverse,
  msgId,
}: {
  progressData: IMsgAttachmentProgress;
  metaData?: IAttachmentData;
  reverse?: boolean;
  msgId: string;
}) {
  const type = useMemo(
    () =>
      metaData
        ? metaData.type === "image/jpeg" ||
          metaData.type === "image/jpg" ||
          metaData.type === "image/png"
          ? "image"
          : "file"
        : "incomplete",
    [metaData]
  );

  const hiddenDownloadLink = useRef<HTMLAnchorElement>(null);
  return (
    <div
      style={reverse ? { flexDirection: "row-reverse" } : {}}
      className={classes.container}
    >
      {failed || !pending ? (
        <>
          {failed && (
            <div
              style={reverse ? { flexDirection: "row-reverse" } : {}}
              className={classes.failed}
            >
              <BiError />
              Upload failed
            </div>
          )}
          {!pending && (
            <>
              {
                {
                  image: <h1>Image</h1>,
                  file: (
                    <>
                      <a
                        download
                        href={`${baseURL}/api/attachment/download/${msgId}`}
                        ref={hiddenDownloadLink}
                      />
                      <button
                        aria-label="Download attachment"
                        name="Download attachment"
                        onClick={() => hiddenDownloadLink.current?.click()}
                        style={reverse ? { flexDirection: "row-reverse" } : {}}
                        className={classes.download}
                      >
                        <AiOutlineDownload />
                        Attachment
                      </button>
                    </>
                  ),
                  incomplete: <h1>Incomplete</h1>,
                }[type]
              }
            </>
          )}
        </>
      ) : (
        <ProgressBar ratio={ratio} />
      )}
    </div>
  );
}
