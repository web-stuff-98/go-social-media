import { useLocation } from "react-router-dom";
import { useContext, createContext, useReducer, useEffect } from "react";
import type { ReactNode } from "react";
import { IDimensions } from "../interfaces/GeneralInterfaces";

type LayoutContainerMode = "Modal" | "Feed" | "Full";

const initialState: State = {
  darkMode: true,
  dimensions: { width: 0, height: 0 },
  containerMode: "Full",
  isMobile: false,
};

function lerp(value1: number, value2: number, amount: number) {
  amount = amount < 0 ? 0 : amount;
  amount = amount > 1 ? 1 : amount;
  return value1 + (value2 - value1) * amount;
}
const normalize = (val: number, max: number, min: number) =>
  (val - min) / (max - min);

type State = {
  darkMode: boolean;
  dimensions: IDimensions;
  containerMode: LayoutContainerMode;
  isMobile: boolean;
};

const InterfaceContext = createContext<{
  state: State;
  dispatch: (action: Partial<State>) => void;
}>({
  state: initialState,
  dispatch: () => {},
});

const InterfaceReducer = (state: State, action: Partial<State>) => ({
  ...state,
  ...action,
});

export const InterfaceProvider = ({ children }: { children: ReactNode }) => {
  const location = useLocation();

  const [state, dispatch] = useReducer(InterfaceReducer, initialState);

  useEffect(() => {
    const handleResize = () => {
      //Fine control over horizontal whitespace...
      const lo = 820;
      const hi = 1280;
      const a =
        (Math.min(hi, Math.max(window.innerWidth, lo)) - lo) / (hi - lo);
      const v = lerp(
        window.innerWidth / 6 / 2,
        window.innerWidth / 1.8 / 2,
        Math.pow(a, 0.5)
      );
      let squareness =
        window.innerWidth > window.innerHeight
          ? window.innerHeight / window.innerWidth
          : window.innerWidth / window.innerHeight;
      squareness = Math.min(Math.max(0, squareness), 1);
      squareness = normalize(squareness, 1, 0.2);
      squareness *= squareness *= squareness;
      document.documentElement.style.setProperty(
        "--horizontal-whitespace",
        `${lerp(window.innerWidth < lo ? 0 : v, 0, squareness)}px`
      );
      dispatch({
        dimensions: { width: window.innerWidth, height: window.innerHeight },
        isMobile: window.innerWidth < lo,
      });
    };
    handleResize();
    window.addEventListener("resize", handleResize);
    const handleDetectDarkmode = (event: MediaQueryListEvent) =>
      dispatch({ darkMode: event?.matches ? true : false });
    window
      .matchMedia("(prefers-color-scheme: dark)")
      .addEventListener("change", handleDetectDarkmode);
    return () => {
      window.removeEventListener("resize", handleResize);
      window
        .matchMedia("(prefers-color-scheme: dark)")
        .removeEventListener("change", handleDetectDarkmode);
    };
  }, []);

  useEffect(() => {
    if (state.darkMode) document.body.classList.add("darkMode");
    else document.body.classList.remove("darkMode");
  }, [state.darkMode]);

  useEffect(() => {
    if (!location) return;
    dispatch({
      containerMode:
        location.pathname.includes("/editor") ||
        location.pathname.includes("/blog") ||
        location.pathname.includes("/post")
          ? "Feed"
          : "Modal",
    });
  }, [location]);

  return (
    <InterfaceContext.Provider value={{ state, dispatch }}>
      {children}
    </InterfaceContext.Provider>
  );
};

export const useInterface = () => useContext(InterfaceContext);
