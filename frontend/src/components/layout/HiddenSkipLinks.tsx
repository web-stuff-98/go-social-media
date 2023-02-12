import { useContext, createContext, useMemo } from "react";
import classes from "../../styles/components/Layout.module.scss";
import type { ReactNode } from "react";
import { useLocation } from "react-router-dom";

/*
    Hidden skip links for screen readers, makes the site easier to navigate using
    keyboard only... skip links are controlled from the Page component the router
    is using.
*/

export type HiddenSkipLink = {
  href: string;
  name: string;
};

const defaultLinks: HiddenSkipLink[] = [
  {
    href: "#nav",
    name: "navigation bar",
  },
  {
    href: "#main",
    name: "main content",
  },
];

export const HiddenSkipLinksContext = createContext<{
  hiddenSkipLinks: HiddenSkipLink[];
}>({
  hiddenSkipLinks: defaultLinks,
});

export function HiddenSkipLinksProvider({ children }: { children: ReactNode }) {
  const { pathname } = useLocation();

  const hiddenSkipLinks = useMemo<HiddenSkipLink[]>(() => {
    let extraLinks: HiddenSkipLink[] = [];

    if (pathname.includes("/blog/")) {
      extraLinks = [
        {
          href: "#aside",
          name: "Aside content",
        },
      ];
    }

    return [...defaultLinks, ...extraLinks];
  }, [pathname]);

  return (
    <HiddenSkipLinksContext.Provider value={{ hiddenSkipLinks }}>
      <ul
        className={classes.accessibilityLinks}
        aria-label="Navigation accessibility links"
        tabIndex={0}
      >
        {hiddenSkipLinks.map((link) => (
          <li>
            <a href={link.href}>{link.name}</a>
          </li>
        ))}
      </ul>
      {children}
    </HiddenSkipLinksContext.Provider>
  );
}

const useHiddenSkipLinks = () => useContext(HiddenSkipLinksContext);
export default useHiddenSkipLinks;
