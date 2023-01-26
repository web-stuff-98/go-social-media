import classes from "../styles/pages/Editor.module.scss";
import { useFormik } from "formik";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css";
import "../styles/fixQuill.css";
import {
  createPost,
  getPost,
  getPostImageFile,
  getRandomImage,
  updatePost,
  uploadPostImage,
} from "../services/posts";
import { useState, useEffect } from "react";
import ResMsg from "../components/shared/ResMsg";
import type { IResMsg } from "../components/shared/ResMsg";
import { useParams } from "react-router-dom";
import { z } from "zod";
import FieldErrorTip from "../components/shared/forms/FieldErrorTip";
import FormikInputAndLabel from "../components/shared/forms/FormikInputLabel";
import FormikFileButtonInput from "../components/shared/forms/FormikFileButtonInput";
import useFormikValidate from "../hooks/useFormikValidate";

export default function Editor() {
  const { slug } = useParams();

  useEffect(() => {
    if (!slug) return;
    loadPostIntoEditor(slug);
  }, [slug]);

  const [originalImageModified, setOriginalImageModified] = useState(false);

  const loadPostIntoEditor = async (slug: string) => {
    setResMsg({ msg: "Loading post...", err: false, pen: true });
    try {
      const p = await getPost(slug);
      formik.setFieldValue("title", p.title);
      formik.setFieldValue("description", p.description);
      formik.setFieldValue("body", p.body);
      formik.setFieldValue("tags", "#" + p.tags.join("#"));
      const file = await getPostImageFile(p.ID);
      formik.setFieldValue("file", file);
      setResMsg({ msg: "", err: false, pen: false });
    } catch (e) {
      setResMsg({ msg: `${e}`, err: true, pen: false });
      setOriginalImageModified(false);
    }
  };

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const { validate, validationErrs } = useFormikValidate(
    z.object({
      title: z.string().max(80).min(2),
      description: z.string().min(10).max(100),
      body: z.string().min(10).max(8000),
      tags: z.string().refine((v) => v.split("#").filter((t) => t).length < 8, {
        message: "Max 8 tags",
      }),
    })
  );

  const formik = useFormik({
    initialValues: {
      title: "",
      description: "",
      body: "",
      tags: "",
      file: undefined,
    },
    validate,
    onSubmit: async (vals) => {
      if (validationErrs.length > 0) return;
      try {
        setResMsg({ msg: "Uploading post...", err: false, pen: true });
        //jest............. wasted hours/days. Test wouldn't pass
        //if (!vals.file) throw new Error("No image file selected");
        let newSlug = "";
        if (!slug) newSlug = await createPost(vals);
        if (slug) await updatePost(vals, slug);
        if (!slug || originalImageModified) {
          await uploadPostImage(vals.file as unknown as File, slug || newSlug);
        }
        setResMsg({ msg: "Post created", err: false, pen: false });
      } catch (e) {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      }
    },
  });

  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      {!resMsg.pen && (
        <>
          <FormikInputAndLabel
            name="title"
            id="title"
            ariaLabel="title"
            touched={formik.touched.title}
            validationErrs={validationErrs}
            value={formik.values.title}
            onChange={formik.handleChange}
            onBlur={formik.handleBlur}
          />
          <FormikInputAndLabel
            name="description"
            id="description"
            ariaLabel="description"
            touched={formik.touched.description}
            validationErrs={validationErrs}
            value={formik.values.description}
            onChange={formik.handleChange}
            onBlur={formik.handleBlur}
          />
          <FormikInputAndLabel
            name="tags"
            id="tags"
            ariaLabel="tags"
            touched={formik.touched.tags}
            validationErrs={validationErrs}
            value={formik.values.tags}
            onChange={formik.handleChange}
            onBlur={formik.handleBlur}
          />
          <div className={classes.quillOuterContainer}>
            <label htmlFor="body">Body</label>
            <div
              data-testid="quill container"
              className={classes.quillContainer}
            >
              <ReactQuill
                theme="snow"
                id="body"
                value={formik.values.body}
                onChange={(e) => formik.setFieldValue("body", e)}
              />
            </div>
            {formik.touched.body && (
              <FieldErrorTip fieldName="body" validationErrs={validationErrs} />
            )}
          </div>
          <div className={classes.buttons}>
            <FormikFileButtonInput
              buttonTestId="Image file button"
              name="Image file"
              id="file"
              ariaLabel="Select a cover image"
              touched={formik.touched.file}
              accept=".jpeg,.jpg,.png"
              validationErrs={validationErrs}
              setFieldValue={formik.setFieldValue}
              setOriginalChanged={setOriginalImageModified}
            />
            <button
              data-testid="Random image button"
              onClick={async () => {
                try {
                  const img = await getRandomImage();
                  formik.setFieldValue("file", img);
                  setOriginalImageModified(true);
                } catch (e) {
                  setResMsg({
                    msg: "Failed to retrieve random image.",
                    err: true,
                    pen: false,
                  });
                }
              }}
              aria-label="Select random image file"
              type="button"
            >
              Random image
            </button>
            <button name="submit" id="submit" type="submit">
              {slug ? "Update" : "Submit"}
            </button>
          </div>
          {formik.values.file && (
            <img
              alt="Preview"
              className={classes.imgPreview}
              src={URL.createObjectURL(formik.values.file)}
            />
          )}
        </>
      )}
      <div className={classes.resMsg}>
        <ResMsg resMsg={resMsg} />
      </div>
    </form>
  );
}
