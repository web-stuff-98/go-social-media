import { useRef, useState, useEffect } from "react";
import type { ReactNode } from "react";
import classes from "../../styles/components/shared/Dropdown.module.scss";

export type DropdownItem = {
  name: string;
  node: ReactNode;
};

export default function Dropdown({
  items = [
    { name: "A", node: <>A</> },
    { name: "B", node: <>B</> },
    { name: "C", node: <>C</> },
  ],
  index = 0,
  setIndex = () => {},
  noRightBorderRadius,
  noLeftBorderRadius,
  noRightBorder,
  noLeftBorder,
  light,
}: {
  items?: DropdownItem[];
  index?: number;
  setIndex?: (to: number) => void;
  noRightBorderRadius?: boolean;
  noLeftBorderRadius?: boolean;
  noRightBorder?: boolean;
  noLeftBorder?: boolean;
  light?: boolean;
}) {
  const rootItemContainerRef = useRef<HTMLDivElement>(null);
  const [dropdownOpen, setDropdownOpen] = useState(false);

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      setDropdownOpen(false);
    }
  };

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const renderItem = (
    name: string,
    children: ReactNode,
    itemIndex: number,
    list?: boolean
  ) => {
    const isStartItem = itemIndex === 0 || !dropdownOpen;
    const isMiddleItem = itemIndex > 0 && itemIndex < items.length - 1;
    const isEndItem = itemIndex === items.length - 1;

    return (
      <button
        data-testid={"Index " + itemIndex}
        className={light ? "lightButton" : ""}
        key={name}
        type="button"
        role="menuitem"
        style={
          list
            ? {
                position: "absolute",
                top: `${
                  rootItemContainerRef.current?.clientHeight! * itemIndex
                }px`,
                ...(isMiddleItem
                  ? {
                      borderTop: "none",
                      borderBottom: "none",
                      borderRadius: "0",
                      height: `${
                        rootItemContainerRef.current?.clientHeight! + 2
                      }px`,
                    }
                  : {
                      ...(isEndItem
                        ? {
                            borderTop: "none",
                            borderTopLeftRadius: "0",
                            borderTopRightRadius: "0",
                          }
                        : {}),
                    }),
              }
            : {
                ...(dropdownOpen && isStartItem
                  ? {
                      borderBottomLeftRadius: "0",
                      borderBottomRightRadius: "0",
                    }
                  : {}),
                ...(isStartItem
                  ? {
                      ...(noLeftBorderRadius
                        ? {
                            borderTopLeftRadius: "0",
                            borderBottomLeftRadius: "0",
                          }
                        : {}),
                      ...(noRightBorderRadius
                        ? {
                            borderTopRightRadius: "0",
                            borderBottomRightRadius: "0",
                          }
                        : {}),
                      ...(noLeftBorder
                        ? {
                            borderLeft: "none",
                          }
                        : {}),
                      ...(noRightBorder
                        ? {
                            borderRight: "none",
                          }
                        : {}),
                    }
                  : {}),
              }
        }
        onClick={() => {
          if (!dropdownOpen) {
            setDropdownOpen(true);
          } else {
            setIndex(itemIndex);
            setDropdownOpen(false);
          }
        }}
        aria-label={dropdownOpen ? `Dropdown menu with current item as ${name}, press enter to open` : name}
      >
        {children}
      </button>
    );
  };

  return (
    <div
      onMouseLeave={() => setDropdownOpen(false)}
      className={classes.container}
      role="menu"
    >
      <div className={classes.inner} ref={rootItemContainerRef}>
        {renderItem(
          items[dropdownOpen ? 0 : index].name,
          items[dropdownOpen ? 0 : index].node,
          dropdownOpen ? 0 : index
        )}
        {dropdownOpen &&
          items
            .slice(1, items.length)
            .map((item, index) =>
              renderItem(item.name, item.node, index + 1, true)
            )}
      </div>
    </div>
  );
}
