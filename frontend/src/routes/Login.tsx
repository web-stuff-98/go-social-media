import classes from "../styles/LoginRegister.module.scss";
import formClasses from "../styles/FormClasses.module.scss";
import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { useAuth } from "../context/AuthContext";
import ResMsg, { IResMsg } from "../components/ResMsg";

export default function Login() {
  const { login } = useAuth();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      setResMsg({ msg: "Logging in", err: false, pen: true });
      await login(username, password);
      setResMsg({ msg: "", err: false, pen: false });
    } catch (e) {
      setResMsg({ msg: `${e}`, err: true, pen: false });
    }
  };

  return (
    <form onSubmit={handleSubmit} className={classes.container}>
      <div className={formClasses.inputLabelWrapper}>
        <label htmlFor="username">Username</label>
        <input
          id="username"
          name="username"
          value={username}
          onChange={(e: ChangeEvent<HTMLInputElement>) =>
            setUsername(e.target.value)
          }
          type="text"
          required
        />
      </div>
      <div className={formClasses.inputLabelWrapper}>
        <label htmlFor="password">Password</label>
        <input
          id="password"
          name="password"
          value={password}
          onChange={(e: ChangeEvent<HTMLInputElement>) =>
            setPassword(e.target.value)
          }
          type="password"
          required
        />
        <button type="submit">Login</button>
      </div>
      <a href="/policy">
        If you login you agree to the privacy / cookies policy.
      </a>
      <ResMsg resMsg={resMsg} />
    </form>
  );
}
