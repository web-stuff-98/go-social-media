import type { ReactNode, CSSProperties } from "react";
import { IconType } from "react-icons/lib";
import classes from "../../styles/components/shared/IconBtn.module.scss";

const IconBtn = ({
  testid,
  Icon,
  children,
  name,
  ariaLabel,
  disabled,
  style,
  svgStyle = {},
  onClick = () => {},
  type = "button",
  ...props
}: {
  testid?: string;
  Icon: IconType;
  children?: ReactNode;
  name: string;
  ariaLabel: string;
  style?: CSSProperties;
  svgStyle?: CSSProperties;
  onClick?: Function;
  disabled?: boolean;
  type?: string;
}) => (
  <button
    {...props}
    style={{
      ...style,
      ...(!disabled ?? false
        ? { cursor: "pointer" }
        : { filter: "opacity(0.5)", pointerEvents: "none" }),
    }}
    type="button"
    name={name}
    aria-label={ariaLabel}
    className={classes.container}
    onClick={() => onClick()}
    data-testid={testid}
  >
    <Icon
      style={{
        ...svgStyle,
        ...(style && style.color ? { fill: style.color } : {}),
      }}
      className={classes.icon}
    />
    {children && children}
  </button>
);
export default IconBtn;
