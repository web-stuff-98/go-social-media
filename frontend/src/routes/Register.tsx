import classes from "../styles/LoginRegister.module.scss";
import { useState } from "react";
import { useAuth } from "../context/AuthContext";
import ResMsg from "../components/shared/ResMsg";
import { z } from "zod";
import { useFormik } from "formik";
import FormikInputAndLabel from "../components/shared/forms/FormikInputLabel";
import useFormikValidate from "../hooks/useFormikValidate";
import { IResMsg } from "../interfaces/GeneralInterfaces";
import { useNavigate } from "react-router-dom";

export default function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const { validate, validationErrs } = useFormikValidate(
    z.object({
      username: z.string().max(16).min(2),
      password: z.string().min(2).max(100),
    })
  );

  const formik = useFormik({
    initialValues: {
      username: "",
      password: "",
    },
    validate,
    onSubmit: async (vals) => {
      if (validationErrs.length > 0) return;
      try {
        setResMsg({ msg: "Creating account...", err: false, pen: true });
        register(vals.username, vals.password);
        setResMsg({ msg: "", err: false, pen: false });
        navigate("/blog/1");
      } catch (e) {
        setResMsg({ msg: `${e}`, err: true, pen: false });
      }
    },
  });

  return (
    <form onSubmit={formik.handleSubmit} className={classes.container}>
      <FormikInputAndLabel
        autoFocus
        name="username"
        id="username"
        ariaLabel="Username"
        value={formik.values.username}
        touched={formik.touched.username}
        validationErrs={validationErrs}
        onChange={formik.handleChange}
        onBlur={formik.handleBlur}
      />
      <FormikInputAndLabel
        password
        name="password"
        id="password"
        ariaLabel="Password"
        value={formik.values.password}
        touched={formik.touched.password}
        validationErrs={validationErrs}
        onChange={formik.handleChange}
        onBlur={formik.handleBlur}
      />
      <a href="/policy">
        If you register you agree to the privacy / cookies policy.
      </a>
      <button type="submit">Create account</button>
      <ResMsg resMsg={resMsg} />
    </form>
  );
}
