import classes from "../styles/components/Attachment.module.scss";
import ProgressBar from "./ProgressBar";
import { useRef } from "react";
import { BiError } from "react-icons/bi";
import { AiOutlineDownload } from "react-icons/ai";
import { baseURL } from "../services/makeRequest";

export interface IMsgAttachmentProgress {
  ratio: number;
  failed: boolean;
  pending: boolean;
}

export default function Attachment({
  progressData: { failed, ratio, pending },
  reverse,
  msgId,
}: {
  progressData: IMsgAttachmentProgress;
  reverse?: boolean;
  msgId: string;
}) {
  const hiddenDownloadLink = useRef<HTMLAnchorElement>(null);
  return (
    <div className={classes.container}>
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
              <a
                download
                href={`${baseURL}/api/attachment/${msgId}`}
                ref={hiddenDownloadLink}
              />
              <button
                aria-label="Download attachment"
                name="Download attachment"
                onClick={() => hiddenDownloadLink.current?.click()}
                style={reverse ? { flexDirection: "row-reverse" } : {}}
                className={classes.complete}
              >
                <AiOutlineDownload />
                Attachment
              </button>
            </>
          )}
        </>
      ) : (
        <ProgressBar ratio={ratio} />
      )}
    </div>
  );
}
