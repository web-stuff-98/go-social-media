import { useFormik } from "formik";
import classes from "../../../styles/components/chat/RoomEditor.module.scss";
import formClasses from "../../../styles/FormClasses.module.scss";
import ResMsg, { IResMsg } from "../../ResMsg";
import { useState, useRef } from "react";
import type { ChangeEvent } from "react";
import { createRoom, uploadRoomImage } from "../../../services/rooms";
import axios from "axios";
import { IRoomCard } from "./Rooms";

export default function RoomEditor() {
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const formik = useFormik({
    initialValues: { name: "", image: null },
    onSubmit: async (vals) => {
      try {
        setResMsg({ msg: "", err: false, pen: true });
        const id = await createRoom(vals as Pick<IRoomCard, "name">);
        if (vals.image) {
          await uploadRoomImage(vals.image, id);
        }
        setResMsg({ msg: "Room created", err: false, pen: false });
      } catch (e) {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      }
    },
  });

  const randomImage = async () => {
    const res = await axios({
      url: "https://picsum.photos/1000/300",
      responseType: "arraybuffer",
    });
    const file = new File([res.data], "image.jpg", { type: "image/jpeg" });
    formik.setFieldValue("image", file);
  };

  const handleImageInput = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    formik.setFieldValue("image", file);
  };

  const imageInputRef = useRef<HTMLInputElement>(null);
  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      <div className={formClasses.inputLabelWrapper}>
        <label htmlFor="name">Room name</label>
        <input
          onChange={formik.handleChange}
          value={formik.values.name}
          name="name"
          id="name"
          type="text"
        />
      </div>
      <div className={formClasses.inputLabelWrapper}>
        <input
          onChange={handleImageInput}
          ref={imageInputRef}
          name="image"
          id="image"
          type="file"
          accept=".jpeg,.jpg,.png"
        />
        <button
          onClick={() => imageInputRef.current?.click()}
          name="Select image"
          aria-label="Select image"
          type="button"
        >
          Select image
        </button>
      </div>
      <button
        name="Random image"
        onClick={() =>
          randomImage().catch((e) =>
            setResMsg({ msg: `${e}`, err: true, pen: false })
          )
        }
        aria-label="Random image"
        type="button"
      >
        Random image
      </button>
      <button type="submit">Create</button>
      {formik.values.image && (
        <img
          alt="Preview"
          className={classes.imgPreview}
          src={URL.createObjectURL(formik.values.image)}
        />
      )}
      <ResMsg resMsg={resMsg} />
    </form>
  );
}
