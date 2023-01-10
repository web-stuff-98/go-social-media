import { FaSearch } from "react-icons/fa";
import classes from "../../../styles/components/chat/Rooms.module.scss";
import IconBtn from "../../IconBtn";

export default function Rooms() {
  return (
    <div className={classes.container}>
      <div className={classes.rooms}></div>
      <form className={classes.searchContainer}>
        <input type="text" placeholder="Search rooms..."/>
        <IconBtn Icon={FaSearch} ariaLabel="Search rooms" name="Search rooms" />
      </form>
    </div>
  );
}
