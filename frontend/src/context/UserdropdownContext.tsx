import {
  useState,
  useContext,
  createContext,
  useEffect,
  useCallback,
  useRef,
} from "react";
import type { ReactNode, ChangeEvent, FormEvent } from "react";
import { MdError, MdSend } from "react-icons/md";
import { BsFillChatRightFill } from "react-icons/bs";
import useScrollbarSize from "react-scrollbar-size";
import useSocket from "./SocketContext";
import classes from "../styles/components/shared/Userdropdown.module.scss";

const UserdropdownContext = createContext<{
  clickPos: { left: string; top: string };
  openUserdropdown: (uid: string) => void;
}>({
  clickPos: { left: "0", top: "0" },
  openUserdropdown: () => {},
});

type UserMenuSection = "Menu" | "DirectMessage";

export function UserdropdownProvider({ children }: { children: ReactNode }) {
  const { width: scrollbarWidth } = useScrollbarSize();
  const { sendIfPossible } = useSocket();

  const containerRef = useRef<HTMLDivElement>(null);
  const [uid, setUid] = useState("");
  const [clickPos, setClickPos] = useState({ left: "0", top: "0" });
  const [cursorInside, setCursorInside] = useState(false);
  const [messageInput, setMessageInput] = useState("");
  const [err, setErr] = useState("");
  const [section, setSection] = useState<UserMenuSection>("Menu");

  const openUserdropdown = (uid: string) => {
    setUid(uid);
  };

  const privateMessageSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    sendIfPossible(
      JSON.stringify({
        event_type: "PRIVATE_MESSAGE",
        content: messageInput,
        recipient_id: uid,
      })
    );
    setMessageInput("");
    closeUserDropdown();
  };

  const closeUserDropdown = () => {
    setUid("");
    setCursorInside(false);
    setSection("Menu");
    setErr("");
  };

  const adjust = useCallback(
    (delay?: boolean) => {
      if (delay) setTimeout(() => internal());
      else internal();
      function internal() {
        if (!containerRef.current) throw new Error("NO CONTAINER REF!!!");
        const leftClickPos = Number(clickPos.left.replace("px", ""));
        const containerRightEndPos =
          leftClickPos + containerRef.current?.clientWidth;
        const padPx = 3 + scrollbarWidth;
        if (containerRightEndPos + padPx > window.innerWidth) {
          setClickPos({
            left: `${
              leftClickPos -
              Math.abs(window.innerWidth - containerRightEndPos - padPx)
            }px`,
            top: clickPos.top,
          });
        }
      }
    },
    [clickPos]
  );
  useEffect(() => {
    if (containerRef.current) adjust();
  }, [uid]);

  const clickedWhileOutside = useCallback((e: MouseEvent) => {
    closeUserDropdown();
    setClickPos({
      left: `${e.clientX}px`,
      top: `${e.clientY}px`,
    });
  }, []);

  useEffect(() => {
    if (uid) adjust(true);
  }, [section]);

  useEffect(() => {
    if (!cursorInside)
      window.addEventListener("mousedown", clickedWhileOutside);
    else window.removeEventListener("mousedown", clickedWhileOutside);
    return () => window.removeEventListener("mousedown", clickedWhileOutside);
  }, [cursorInside]);

  return (
    <UserdropdownContext.Provider value={{ clickPos, openUserdropdown }}>
      {uid && (
        <div
          ref={containerRef}
          onMouseEnter={() => setCursorInside(true)}
          onMouseLeave={() => setCursorInside(false)}
          aria-label="User dropdown"
          style={{ left: clickPos.left, top: clickPos.top, zIndex: 100 }}
          className={classes.container}
        >
          {err ? (
            <div className={classes.errContainer}>
              <MdError />
              {err}
            </div>
          ) : (
            <>
              {section === "Menu" && (
                <div className={classes.menu}>
                  <button
                    aria-label="Message"
                    onClick={() => setSection("DirectMessage")}
                  >
                    <BsFillChatRightFill />
                    Chat
                  </button>
                </div>
              )}
              {section === "DirectMessage" && (
                <form
                  className={classes.messageForm}
                  onSubmit={privateMessageSubmit}
                >
                  <input
                    autoFocus
                    value={messageInput}
                    onChange={(e: ChangeEvent<HTMLInputElement>) =>
                      setMessageInput(e.target.value)
                    }
                    placeholder="Direct message..."
                    required
                    type="text"
                  />
                  <button aria-label="Send direct message">
                    <MdSend />
                  </button>
                </form>
              )}
            </>
          )}
        </div>
      )}
      {children}
    </UserdropdownContext.Provider>
  );
}

export const useUserdropdown = () => useContext(UserdropdownContext);
