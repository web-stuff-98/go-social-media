import classes from "../styles/components/Attachment.module.scss";
import ProgressBar from "./ProgressBar";

import { BiError } from "react-icons/bi";
import { AiOutlineDownload } from "react-icons/ai";

export interface IMsgAttachmentData {
  ratio: number;
  failed: boolean;
  pending: boolean;
}

export default function Attachment({
  data: { failed, ratio, pending },
  reverse,
}: {
  data: IMsgAttachmentData;
  reverse?: boolean;
}) {
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
            <div
              style={reverse ? { flexDirection: "row-reverse" } : {}}
              className={classes.complete}
            >
              <AiOutlineDownload />
              Attachment
            </div>
          )}
        </>
      ) : (
        <ProgressBar ratio={ratio} />
      )}
    </div>
  );
}
