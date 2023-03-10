import formClasses from "../../styles/FormClasses.module.scss";
import { ImSpinner8 } from "react-icons/im";
import { IResMsg } from "../../interfaces/GeneralInterfaces";

export default function ResMsg({
  resMsg,
  large,
}: {
  resMsg: IResMsg;
  large?: boolean;
}) {
  return (
    <>
      {(resMsg.msg || resMsg.pen) && (
        <div
          aria-label="Loading"
          aria-live="assertive"
          style={resMsg.pen ? { paddingTop: "var(--padding)" } : {}}
          className={resMsg.err ? formClasses.resMsgErr : formClasses.resMsg}
        >
          {resMsg.pen && (
            <ImSpinner8
              data-testid="Loading spinner"
              style={large ? { width: "2rem", height: "2rem" } : {}}
              className={formClasses.loadingSpinner}
            />
          )}
          {resMsg.msg}
        </div>
      )}
    </>
  );
}
