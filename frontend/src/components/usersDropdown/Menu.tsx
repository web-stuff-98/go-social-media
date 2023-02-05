import { BiDoorOpen } from "react-icons/bi";
import { BsFillChatRightFill } from "react-icons/bs";
import { IoBan } from "react-icons/io5";
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
        aria-label="Message user"
        onClick={() => setDropdownSectionTo("DirectMessage")}
      >
        <BsFillChatRightFill />
        Direct message
      </button>
      <button
        aria-label="Invite to room"
        onClick={() => setDropdownSectionTo("InviteToRoom")}
      >
        <BiDoorOpen />
        Invite to room
      </button>
      <button
        aria-label="Ban from room"
        onClick={() => setDropdownSectionTo("BanFromRoom")}
      >
        <IoBan />
      </button>
    </div>
  );
}
