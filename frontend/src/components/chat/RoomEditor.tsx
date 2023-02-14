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
import FormikInputAndLabel from "../shared/forms/FormikInputLabel";
import FormikFileButtonInput from "../shared/forms/FormikFileButtonInput";
import useFormikValidate from "../../hooks/useFormikValidate";
import { IResMsg } from "../../interfaces/GeneralInterfaces";
import { IRoomCard } from "../../interfaces/ChatInterfaces";
import Toggle from "../shared/Toggle";
import useChat from "../../context/ChatContext";

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
      formik.setFieldValue("private", r.private);
      setImgURL(URL.createObjectURL(b));
      setResMsg({ msg: "", err: false, pen: false });
    } catch (e) {
      //If its a room image error that means the room has no image, so its
      //just a not found error
      setResMsg({ msg: "", err: `${e}` !== "Room image error", pen: false });
    }
  };

  useEffect(() => {
    if (!editRoomId) return;
    const controller = new AbortController();
    loadRoom();
    return () => {
      controller.abort();
    };
    // eslint-disable-next-line
  }, [editRoomId]);

  const { validate, validationErrs } = useFormikValidate(
    z.object({
      name: z.string().max(16).min(2),
    })
  );

  const formik = useFormik({
    initialValues: { name: "", image: null, private: false },
    validate,
    onSubmit: async (vals: { name: string; image?: any; private: boolean }) => {
      try {
        setResMsg({ msg: "", pen: true, err: false });
        if (validationErrs.length > 0) return;
        await handleCreateUpdateRoom(
          vals,
          originalImageChanged,
        );
        setResMsg({ msg: "", pen: false, err: false });
      } catch (e) {
        setResMsg({ msg: `${e}`, pen: false, err: true });
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
            name="name"
            id="name"
            ariaLabel="Room name"
            validationErrs={validationErrs}
            value={formik.values.name}
            onChange={formik.handleChange}
            onBlur={formik.handleBlur}
          />
          <FormikFileButtonInput
            buttonTestId="Image file button"
            name="room image"
            id="image"
            ariaControls="room image"
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
          <Toggle
            label="Private"
            toggledOn={formik.values.private}
            setToggledOn={(to: boolean) => formik.setFieldValue("private", to)}
          />
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
            marginBottom: "var(--padding)",
          }}
        >
          <ResMsg resMsg={resMsg} />
        </div>
      )}
    </form>
  );
}
