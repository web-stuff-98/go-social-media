import formClasses from "../styles/FormClasses.module.scss";

import { ImSpinner8 } from "react-icons/im";

export interface IResMsg {
  msg: string;
  err: boolean;
  pen: boolean;
}

export default function ResMsg({
  resMsg,
  large
}: {
  resMsg: IResMsg;
  large?: boolean;
}) {
  return (
    <>
      {(resMsg.msg || resMsg.pen) && (
        <div
          style={resMsg.pen ? { paddingTop: "var(--padding)" } : {}}
          className={resMsg.err ? formClasses.resMsgErr : formClasses.resMsg}
        >
          {resMsg.pen && <ImSpinner8 style={large ? { width: "2rem", height:"2rem"} : {}} className={formClasses.loadingSpinner} />}
          {resMsg.msg}
        </div>
      )}
    </>
  );
}
