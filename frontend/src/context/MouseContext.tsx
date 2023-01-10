import type { ReactNode } from "react";
import { createContext, useContext, useState, useEffect } from "react";

import type { IPosition } from "../context/InterfaceContext";

const MouseContext = createContext<IPosition | undefined>(undefined);

export function MouseProvider({ children }: { children: ReactNode }) {
  const [mousePos, setMousePos] = useState<IPosition>({ left: 0, top: 0 });
  const detectMousePos = (e: MouseEvent) =>
    setMousePos({ left: e.clientX, top: e.clientY });
  useEffect(() => {
    window.addEventListener("mousemove", detectMousePos);
    return () => window.removeEventListener("mousemove", detectMousePos);
  }, []);
  return (
    <MouseContext.Provider value={mousePos}>{children}</MouseContext.Provider>
  );
}

export const useMouse = () => useContext(MouseContext);
