import { useFormik } from "formik";
import classes from "../../styles/components/chat/RoomEditor.module.scss";
import formClasses from "../../styles/FormClasses.module.scss"
import ResMsg, { IResMsg } from "../shared/ResMsg";
import { useState, useRef, useEffect } from "react";
import type { ChangeEvent } from "react";
import {
  createRoom,
  getRoom,
  getRoomImageAsBlob,
  updateRoom,
  uploadRoomImage,
} from "../../services/rooms";
import axios from "axios";
import { IRoomCard } from "./Rooms";
import { z } from "zod";
import FieldErrorTip from "../shared/FieldErrorTip";
import { useChat } from "./Chat";

export default function RoomEditor() {
  const { editRoomId } = useChat();

  useEffect(() => {
    if (!editRoomId) return;
    loadRoom();
  }, [editRoomId]);

  const [originalImageChanged, setOriginalImageChanged] = useState(false);
  const [imgURL, setImgURL] = useState("");
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const loadRoom = () => {
    setResMsg({ msg: "", err: false, pen: true });
    getRoom(editRoomId)
      .then((r: IRoomCard) => {
        formik.setFieldValue("name", r.name);
        getRoomImageAsBlob(editRoomId)
          .then((b) => {
            formik.setFieldValue("image", b);
            setImgURL(URL.createObjectURL(b));
          })
          .catch(() => {});
        setResMsg({ msg: "", err: false, pen: false });
      })
      .catch((e) => {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      });
  };

  const Schema = z.object({
    name: z.string().max(16).min(2),
  });

  const [validationErrs, setValidationErrs] = useState<any[]>([]);

  const formik = useFormik({
    initialValues: { name: "", image: null },
    validate: (values) => {
      if (!Schema) return;
      try {
        Schema.parse(values);
        setValidationErrs([]);
      } catch (error: any) {
        setValidationErrs(error.issues);
      }
    },
    onSubmit: async (vals) => {
      if (validationErrs.length > 0) return;
      try {
        setResMsg({ msg: "", err: false, pen: true });
        let id: string;
        if (editRoomId) {
          id = editRoomId;
          await updateRoom({
            ...(vals as Pick<IRoomCard, "name">),
            ID: editRoomId,
          });
        } else {
          id = await createRoom(vals as Pick<IRoomCard, "name">);
        }
        if (
          (vals.image && !editRoomId) ||
          (editRoomId && originalImageChanged && vals.image)
        ) {
          await uploadRoomImage(vals.image, id);
        }
        setResMsg({
          msg: editRoomId ? "Room updated" : "Room created",
          err: false,
          pen: false,
        });
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
    setImgURL(URL.createObjectURL(file));
    setOriginalImageChanged(true);
  };

  const handleImageInput = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    formik.setFieldValue("image", file);
    setImgURL(URL.createObjectURL(file));
    setOriginalImageChanged(true);
  };

  const imageInputRef = useRef<HTMLInputElement>(null);
  const imgRef = useRef<HTMLImageElement>(null);
  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      {!resMsg.pen && (
        <>
          <div className={formClasses.inputLabelWrapper}>
            <label htmlFor="name">Room name</label>
            <input
              onChange={formik.handleChange}
              value={formik.values.name}
              name="name"
              id="name"
              type="text"
            />
            {formik.touched.name && <FieldErrorTip fieldName="name" validationErrs={validationErrs} />}
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
              ref={imgRef}
              alt="Preview"
              className={classes.imgPreview}
              src={imgURL}
            />
          )}
        </>
      )}
      {(resMsg.pen || resMsg.msg) && (
        <div
          style={{
            width: `${imgRef.current ? `${imgRef.current.width}px` : "12rem"}`,
          }}
        >
          <ResMsg resMsg={resMsg} />
        </div>
      )}
    </form>
  );
}
