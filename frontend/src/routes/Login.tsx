import classes from "../styles/LoginRegister.module.scss";
import { useState } from "react";
import { useAuth } from "../context/AuthContext";
import ResMsg from "../components/shared/ResMsg";
import { z } from "zod";
import { useFormik } from "formik";
import useFormikValidate from "../hooks/useFormikValidate";
import FormikInputAndLabel from "../components/shared/forms/FormikInputLabel";
import { IResMsg } from "../interfaces/GeneralInterfaces";
import { useNavigate } from "react-router-dom";

export default function Login() {
  const { login } = useAuth();
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
      try {
        setResMsg({ msg: "Logging in", err: false, pen: true });
        login(vals.username, vals.password);
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
        onChange={formik.handleChange}
        onBlur={formik.handleBlur}
        validationErrs={validationErrs}
      />
      <FormikInputAndLabel
        password
        name="password"
        id="password"
        ariaLabel="Password"
        value={formik.values.password}
        touched={formik.touched.password}
        onChange={formik.handleChange}
        onBlur={formik.handleBlur}
        validationErrs={validationErrs}
      />
      <a href="/policy">
        If you login you agree to the privacy / cookies policy.
      </a>
      <button type="submit">Login</button>
      <ResMsg resMsg={resMsg} />
    </form>
  );
}
