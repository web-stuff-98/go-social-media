import classes from "../styles/pages/Editor.module.scss";
import formClasses from "../styles/FormClasses.module.scss";
import { useFormik } from "formik";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css";
import "../styles/fixQuill.css";
import {
  createPost,
  getPost,
  getPostImage,
  getPostImageFile,
  updatePost,
  uploadPostImage,
} from "../services/posts";
import { useState, useRef, useEffect } from "react";
import type { ChangeEvent } from "react";
import ResMsg from "../components/ResMsg";
import type { IResMsg } from "../components/ResMsg";
import axios from "axios";
import { useParams } from "react-router-dom";

import { z } from "zod";
import FieldErrorTip from "../components/FieldErrorTip";

export default function Editor() {
  const { slug } = useParams();

  useEffect(() => {
    if (!slug) return;
    loadPostIntoEditor(slug);
  }, [slug]);

  const [originalImageModified, setOriginalImageModified] = useState(false);

  const loadPostIntoEditor = (slug: string) => {
    setResMsg({ msg: "Loading post...", err: false, pen: true });
    getPost(slug)
      .then((p) => {
        formik.setFieldValue("title", p.title);
        formik.setFieldValue("description", p.description);
        formik.setFieldValue("body", p.body);
        formik.setFieldValue("tags", "#" + p.tags.join("#"));
        getPostImageFile(p.ID)
          .then((file) => {
            formik.setFieldValue("file", file);
            setResMsg({ msg: "", err: false, pen: false });
          })
          .catch((e) => {
            throw new Error(e);
          });
      })
      .catch((e) => {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      })
      .finally(() => {
        setOriginalImageModified(false);
      });
  };

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const [validationErrs, setValidationErrs] = useState<any[]>([]);

  const Schema = z.object({
    title: z.string().max(80).min(2),
    description: z.string().min(10).max(100),
    body: z.string().min(10).max(8000),
    tags: z.string().refine((v) => v.split("#").filter((t) => t).length < 8, {
      message: "Max 8 tags",
    }),
  });

  const formik = useFormik({
    initialValues: {
      title: "",
      description: "",
      body: "",
      tags: "",
      file: undefined,
    },
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
        setResMsg({ msg: "Uploading post...", err: false, pen: true });
        if (!vals.file) throw new Error("No image file selected");
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

  const randomImage = async () => {
    const res = await axios({
      url: "https://picsum.photos/1000/400",
      responseType: "arraybuffer",
    });
    const file = new File([res.data], "image.jpg", { type: "image/jpeg" });
    formik.setFieldValue("file", file);
    setOriginalImageModified(true);
  };

  const handleFileInput = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    if (!e.target.files[0]) return;
    const file = e.target.files[0];
    formik.setFieldValue("file", file);
    setOriginalImageModified(true);
  };

  const fileInputRef = useRef<HTMLInputElement>(null);
  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      {!resMsg.pen && (
        <>
          <div className={formClasses.inputLabelWrapper}>
            <label htmlFor="title">Title</label>
            <input
              name="title"
              id="title"
              aria-label="title"
              value={formik.values.title}
              onChange={formik.handleChange}
              onBlur={formik.handleBlur}
            />
            <FieldErrorTip fieldName="title" validationErrs={validationErrs} />
          </div>
          <div className={formClasses.inputLabelWrapper}>
            <label htmlFor="description">Description</label>
            <input
              name="description"
              id="description"
              aria-label="description"
              value={formik.values.description}
              onChange={formik.handleChange}
              onBlur={formik.handleBlur}
            />
            <FieldErrorTip
              fieldName="description"
              validationErrs={validationErrs}
            />
          </div>
          <div className={formClasses.inputLabelWrapper}>
            <label htmlFor="tags">Tags</label>
            <input
              name="tags"
              id="tags"
              aria-label="tags"
              value={formik.values.tags}
              onChange={formik.handleChange}
              onBlur={formik.handleBlur}
            />
            <FieldErrorTip fieldName="tags" validationErrs={validationErrs} />
          </div>
          <div className={classes.quillOuterContainer}>
            <label htmlFor="body">Body</label>
            <div className={classes.quillContainer}>
              <ReactQuill
                theme="snow"
                id="body"
                value={formik.values.body}
                onChange={(e) => formik.setFieldValue("body", e)}
              />
            </div>
            <FieldErrorTip fieldName="body" validationErrs={validationErrs} />
          </div>
          <input
            name="file"
            id="file"
            onChange={handleFileInput}
            ref={fileInputRef}
            type="file"
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            aria-label="Select image file"
            type="button"
          >
            Select image
          </button>
          <button
            onClick={() => randomImage()}
            aria-label="Select random image file"
            type="button"
          >
            Random image
          </button>
          <button name="submit" id="submit" type="submit">
            {slug ? "Update" : "Submit"}
          </button>
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
