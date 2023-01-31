import { useFormik } from "formik";
import classes from "../../styles/components/chat/RoomEditor.module.scss";
import ResMsg from "../shared/ResMsg";
import { useState, useRef, useEffect } from "react";
import {
  getRandomRoomImage,
  getRoom,
  getRoomImageAsBlob,
} from "../../services/rooms";
import { z } from "zod";
import { useChat } from "./Chat";
import FormikInputAndLabel from "../shared/forms/FormikInputLabel";
import FormikFileButtonInput from "../shared/forms/FormikFileButtonInput";
import useFormikValidate from "../../hooks/useFormikValidate";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
import { IRoomCard } from "../../interfaces/ChatInterfaces";

export default function RoomEditor() {
  const { editRoomId, handleCreateUpdateRoom } = useChat();

  const [originalImageChanged, setOriginalImageChanged] = useState(false);
  const [imgURL, setImgURL] = useState("");
  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const loadRoom = async () => {
    setResMsg({ msg: "", err: false, pen: true });
    try {
      const r: IRoomCard = await getRoom(editRoomId);
      formik.setFieldValue("name", r.name);
      const b: Blob = await getRoomImageAsBlob(editRoomId).catch(() => {
        throw new Error("Room image error");
      });
      formik.setFieldValue("image", b);
      setImgURL(URL.createObjectURL(b));
    } catch (e) {
      //If its a room image error that means the room has no image, so its
      //just a not found error
      setResMsg({ msg: "", err: `${e}` !== "Room image error", pen: false });
    }
  };

  useEffect(() => {
    if (!editRoomId) return;
    loadRoom();
    // eslint-disable-next-line
  }, [editRoomId]);

  const { validate, validationErrs } = useFormikValidate(
    z.object({
      name: z.string().max(16).min(2),
    })
  );

  const formik = useFormik({
    initialValues: { name: "", image: null },
    validate,
    onSubmit: async (vals: { name: string; image?: any }) => {
      try {
        setResMsg({ msg: "", pen: true, err: false });
        handleCreateUpdateRoom(
          vals,
          originalImageChanged,
          validationErrs.length
        );
        setResMsg({ msg: "", pen: false, err: false });
      } catch (e) {
        setResMsg({ msg: `${e}`, pen: false, err: false });
      }
    },
  });

  const randomImage = async () => {
    const file = await getRandomRoomImage();
    formik.setFieldValue("image", file);
    setImgURL(URL.createObjectURL(file));
    setOriginalImageChanged(true);
  };

  const imgRef = useRef<HTMLImageElement>(null);
  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      {!resMsg.pen && (
        <>
          <FormikInputAndLabel
            touched={formik.touched.name}
            name="Room name"
            id="name"
            ariaLabel="Room name"
            validationErrs={validationErrs}
            value={formik.values.name}
            onChange={formik.handleChange}
            onBlur={formik.handleBlur}
          />
          <FormikFileButtonInput
            buttonTestId="Image file button"
            name="Select room image"
            id="image"
            ariaLabel="Select room image"
            accept=".jpeg,.jpeg,.png"
            touched={formik.touched.image}
            validationErrs={validationErrs}
            setFieldValue={formik.setFieldValue}
            setURL={setImgURL}
            setOriginalChanged={setOriginalImageChanged}
          />
          <button
            name="Random image"
            onClick={async () => {
              try {
                await randomImage();
              } catch (e) {
                setResMsg({ msg: `${e}`, err: true, pen: false });
              }
            }}
            aria-label="Random image"
            type="button"
          >
            Random image
          </button>
          <button
            aria-label={editRoomId ? "Update room" : "Create room"}
            name={editRoomId ? "Update room" : "Create room"}
            type="submit"
          >
            {editRoomId ? "Update" : "Create"}
          </button>
          <img
            data-testid="Image preview"
            ref={imgRef}
            alt="Room cover preview"
            style={formik.values.image ? {} : { display: "none" }}
            className={classes.imgPreview}
            src={imgURL}
          />
        </>
      )}
      {(resMsg.pen || resMsg.msg) && (
        <div
          data-testid="ResMsg container"
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
