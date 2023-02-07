import classes from "../../styles/components/chat/Attachment.module.scss";
import ProgressBar from "../shared/ProgressBar";
import { useMemo, useRef } from "react";
import { BiError } from "react-icons/bi";
import { AiOutlineDownload } from "react-icons/ai";
import { baseURL } from "../../services/makeRequest";
import {
  IAttachmentData,
  IMsgAttachmentProgress,
} from "../../interfaces/ChatInterfaces";

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
                  image: (
                    <img
                      alt={metaData?.name}
                      className={classes.image}
                      src={`${baseURL}/api/attachment/download/${msgId}`}
                    />
                  ),
                  file: (
                    <a
                      aria-label="Download attachment"
                      className={classes.download}
                      style={reverse ? { flexDirection: "row-reverse" } : {}}
                      download
                      href={`${baseURL}/api/attachment/download/${msgId}`}
                      ref={hiddenDownloadLink}
                    >
                      <AiOutlineDownload />
                      Attachment
                    </a>
                  ),
                  /*video: (
                    <div className={classes.videoPlayerContainer}>
                      <ReactPlayer
                        width="100%"
                        height="auto"
                        url={`${baseURL}/api/attachment/video/${msgId}`}
                      />
                    </div>
                  ),*/
                  incomplete: <h1>this should never be visible</h1>,
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
