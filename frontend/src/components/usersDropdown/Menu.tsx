import { BiDoorOpen } from "react-icons/bi";
import { BsFillChatRightFill } from "react-icons/bs";
import { FaBan } from "react-icons/fa";
import { UserMenuSection } from "../../context/UserdropdownContext";
import classes from "../../styles/components/shared/Userdropdown.module.scss";

export default function Menu({
  setDropdownSectionTo,
}: {
  setDropdownSectionTo: (section: UserMenuSection) => void;
}) {
  return (
    <div className={classes.menu}>
      <button
        autoFocus
        tabIndex={0}
        aria-label="Message user"
        onClick={() => setDropdownSectionTo("DirectMessage")}
      >
        <BsFillChatRightFill />
        Direct message
      </button>
      <button
        tabIndex={1}
        aria-label="Invite to room"
        onClick={() => setDropdownSectionTo("InviteToRoom")}
      >
        <BiDoorOpen />
        Invite to room
      </button>
      <button
        tabIndex={2}
        aria-label="Ban from room"
        onClick={() => setDropdownSectionTo("BanFromRoom")}
      >
        <FaBan />
        Ban from room
      </button>
    </div>
  );
}
