import {
  useState,
  useContext,
  createContext,
  useEffect,
  useCallback,
  useRef,
} from "react";
import type { ReactNode } from "react";
import { MdError } from "react-icons/md";
import useScrollbarSize from "react-scrollbar-size";
import classes from "../styles/components/shared/Userdropdown.module.scss";
import Menu from "../components/usersDropdown/Menu";
import DirectMessage from "../components/usersDropdown/DirectMessage";
import InviteToRoom from "../components/usersDropdown/InviteToRoom";
import BanFromRoom from "../components/usersDropdown/BanFromRoom";

const UserdropdownContext = createContext<{
  clickPos: { left: string; top: string };
  openUserdropdown: (uid: string) => void;
}>({
  clickPos: { left: "0", top: "0" },
  openUserdropdown: () => {},
});

export type UserMenuSection =
  | "Menu"
  | "DirectMessage"
  | "InviteToRoom"
  | "BanFromRoom";

export function UserdropdownProvider({ children }: { children: ReactNode }) {
  const { width: scrollbarWidth } = useScrollbarSize();

  const containerRef = useRef<HTMLDivElement>(null);
  const [uid, setUid] = useState("");
  const [clickPos, setClickPos] = useState({ left: "0", top: "0" });
  const [cursorInside, setCursorInside] = useState(false);
  const [err, setErr] = useState("");
  const [section, setSection] = useState<UserMenuSection>("Menu");

  const openUserdropdown = (uid: string) => {
    setUid(uid);
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
    // eslint-disable-next-line
    [clickPos]
  );
  useEffect(() => {
    if (containerRef.current) adjust();
    // eslint-disable-next-line
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
    // eslint-disable-next-line
  }, [section]);

  useEffect(() => {
    if (!cursorInside)
      window.addEventListener("mousedown", clickedWhileOutside);
    else window.removeEventListener("mousedown", clickedWhileOutside);
    return () => window.removeEventListener("mousedown", clickedWhileOutside);
    // eslint-disable-next-line
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
            <div aria-label="Error" className={classes.errContainer}>
              <MdError />
              {err}
            </div>
          ) : (
            <>
              {
                {
                  Menu: <Menu setDropdownSectionTo={setSection} />,
                  DirectMessage: (
                    <DirectMessage
                      setDropdownSectionTo={setSection}
                      closeUserDropdown={closeUserDropdown}
                      uid={uid}
                    />
                  ),
                  InviteToRoom: (
                    <InviteToRoom
                      uid={uid}
                      closeUserDropdown={closeUserDropdown}
                    />
                  ),
                  BanFromRoom: (
                    <BanFromRoom
                      uid={uid}
                      closeUserDropdown={closeUserDropdown}
                    />
                  ),
                }[section]
              }
            </>
          )}
        </div>
      )}
      {children}
    </UserdropdownContext.Provider>
  );
}

export const useUserdropdown = () => useContext(UserdropdownContext);
