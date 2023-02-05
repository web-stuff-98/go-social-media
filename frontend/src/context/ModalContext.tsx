import {
  useState,
  useContext,
  createContext,
  useCallback,
  useEffect,
} from "react";
import type { ReactNode } from "react";
import { ImSpinner8 } from "react-icons/im";
import { BiError } from "react-icons/bi";

import classes from "../styles/components/shared/Modal.module.scss";
import { instanceOfResponseMessageData } from "../utils/DetermineSocketEvent";
import useSocket from "./SocketContext";

/**
 *
 * modalType can be either "Confirm", "Message"
 *
 * modalData :
 *  pen = something is pending / loading
 *  err = message is an error
 *  msg = message content
 *  confirmationCallback = the asynchronous promise invoked after confirmation
 *  cancellationCallback = your cancellation function
 *
 * "Message" modal type have an error, a loading spinner
 * or just a message on its own.
 *
 */

interface IModalData {
  pen: boolean;
  err: boolean;
  msg: string;
  confirmationCallback?: Function;
  cancellationCallback?: Function;
}

const defaultModalData = {
  msg: "Hello",
  err: false,
  pen: false,
  confirmationCallback: () => {},
  cancellationCallback: () => {},
};

export const ModalContext = createContext<{
  openModal: (modalType: "Message" | "Confirm", modalData: IModalData) => void;
  closeModal: () => void;
  setData: (data: Partial<IModalData>) => void;
}>({
  openModal: () => {},
  closeModal: () => {},
  setData: () => {},
});

export const ModalProvider = ({ children }: { children: ReactNode }) => {
  const { socket } = useSocket();

  const [modalType, setModalType] = useState<"Message" | "Confirm">("Message");
  const [modalData, setModalData] = useState<IModalData>(defaultModalData);
  const [showModal, setShowModal] = useState(false);

  const openModal = (
    modalType: "Message" | "Confirm",
    modalData: Partial<IModalData>
  ) => {
    setModalData((old) => ({ ...old, ...modalData }));
    setShowModal(true);
    setModalType(modalType);
  };
  const closeModal = () => {
    setShowModal(false);
  };
  const setData = (data: Partial<IModalData>) =>
    setModalData((old) => ({ ...old, ...data }));

  const handleMessage = useCallback((e: MessageEvent) => {
    const data = JSON.parse(e.data);
    if (!data["DATA"]) return;
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfResponseMessageData(e.data)) {
      openModal("Message", {
        msg: data.msg,
        err: data.err,
        pen: false,
      });
    }
    // eslint-disable-next-line
  }, []);

  useEffect(() => {
    socket?.addEventListener("message", handleMessage);
    return () => {
      socket?.removeEventListener("message", handleMessage);
    };
    // eslint-disable-next-line
  }, [socket]);

  return (
    <ModalContext.Provider value={{ openModal, closeModal, setData }}>
      <>
        {showModal && (
          <>
            <div className={classes.backdrop} />
            <div
              onClick={() => {
                if (!modalData.pen) closeModal();
              }}
              className={classes.modalContainer}
            >
              <div className={classes.modal}>
                {/* Confirmation modal */}
                {showModal && modalType === "Confirm" && (
                  <>
                    <b>{modalData.msg}</b>
                    <div className={classes.buttons}>
                      <button
                        aria-label="Cancel"
                        onClick={() => {
                          if (modalData.cancellationCallback)
                            modalData.cancellationCallback();
                          closeModal();
                        }}
                      >
                        Cancel
                      </button>
                      <button
                        aria-label="Confirm"
                        onClick={() => {
                          if (modalData.confirmationCallback)
                            modalData.confirmationCallback();
                          closeModal();
                        }}
                      >
                        Confirm
                      </button>
                    </div>
                  </>
                )}
                {/* Message modal */}
                {showModal && modalType === "Message" && (
                  <>
                    {modalData.pen && (
                      <ImSpinner8 className={classes.spinner} />
                    )}
                    {modalData.err && <BiError className={classes.error} />}
                    <b>{modalData.msg}</b>
                  </>
                )}
              </div>
            </div>
          </>
        )}
      </>
      {children}
    </ModalContext.Provider>
  );
};

export const useModal = () => useContext(ModalContext);
