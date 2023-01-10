import classes from "../../../styles/components/chat/RoomEditor.module.scss";
import formClasses from "../../../styles/FormClasses.module.scss";

export default function RoomEditor() {
  return (
    <form className={classes.container}>
      <div className={formClasses.inputLabelWrapper}>
        <label htmlFor="name">Room name</label>
        <input name="name" id="name" type="text" />
      </div>
      <div className={formClasses.inputLabelWrapper}>
        <input name="image" id="image" type="file" />
        <button name="Select image" aria-label="Select image" type="button">
          Select image
        </button>
      </div>
      <button type="submit">Create</button>
    </form>
  );
}
