import { useRef, useState } from "react";
import type { ChangeEvent } from "react";
import User from "../components/shared/User";
import classes from "../styles/pages/Settings.module.scss";
import { useAuth } from "../context/AuthContext";
import ResMsg from "../components/shared/ResMsg";
import { makeRequest } from "../services/makeRequest";
import ProtectedRoute from "./ProtectedRoute";
import { BsGearWide } from "react-icons/bs";
import { useModal } from "../context/ModalContext";
import { IResMsg } from "../interfaces/GeneralInterfaces";

export default function Settings() {
  const { user, deleteAccount, updateUserState } = useAuth();
  const { openModal, closeModal } = useModal();

  const fileRef = useRef<File>();

  const [resMsg, setResMsg] = useState<IResMsg>({
    msg: "",
    err: false,
    pen: false,
  });

  const handlePfpInput = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    const file = e.target.files![0];
    if (!file) return;
    fileRef.current = file;
    updatePfp(file.name);
  };

  const updatePfp = async (name: string) => {
    try {
      if (!fileRef.current) throw new Error("No image selected");
      openModal("Confirm", {
        err: false,
        pen: false,
        msg: `Are you sure you want to use ${name} as your profile picture?`,
        confirmationCallback: async () => {
          setResMsg({ msg: "", err: false, pen: true });
          try {
            const formData = new FormData();
            formData.append("file", fileRef.current as File, "pfp");
            await makeRequest("/api/account/pfp", {
              method: "POST",
              withCredentials: true,
              data: formData,
            });
            const b64 = await new Promise<string>((resolve, reject) => {
              const fr = new FileReader();
              fr.readAsDataURL(fileRef.current!);
              fr.onloadend = () => resolve(fr.result as string);
              fr.onabort = () => reject("Aborted");
              fr.onerror = () => reject("Error");
            });
            updateUserState({ base64pfp: b64 });
            closeModal();
          } catch (e) {
            openModal("Message", {
              err: true,
              msg: `${e}`,
              pen: false,
            });
            setResMsg({ msg: "", err: false, pen: false });
          }
        },
        cancellationCallback: () => {},
      });
    } catch (e) {
      setResMsg({ msg: `${e}`, err: true, pen: false });
    }
  };

  const hiddenPfpInputRef = useRef<HTMLInputElement>(null);
  return (
    <ProtectedRoute user={user}>
      <form className={classes.container}>
        <div className={classes.heading}>
          <BsGearWide />
          <h1>Settings</h1>
        </div>
        <hr aria-hidden="true" />
        <input
          onChange={handlePfpInput}
          type="file"
          ref={hiddenPfpInputRef}
          accept=".jpeg,.jpg,.png"
        />
        <User
          testid="User"
          uid={user?.ID!}
          user={user}
          onClick={() => hiddenPfpInputRef.current?.click()}
        />
        <p>
          You can click on your profile picture above to select a new image. It
          will update for other users automatically after confirming your
          selection.
        </p>
        <button
          aria-label="Delete account"
          name="Delete account"
          onClick={() =>
            openModal("Confirm", {
              err: false,
              pen: false,
              msg: "Are you sure you want to delete your account?",
              confirmationCallback: async () => {
                try {
                  openModal("Message", {
                    msg: "Deleting account",
                    err: false,
                    pen: true,
                  });
                  deleteAccount();
                  openModal("Message", {
                    msg: "Account deleted",
                    err: false,
                    pen: false,
                  });
                } catch (e) {
                  openModal("Message", {
                    err: true,
                    pen: false,
                    msg: `${e}`,
                  });
                }
              },
              cancellationCallback: () => {},
            })
          }
          className={classes.deleteAccountButton}
          type="button"
        >
          Delete account
        </button>
        <ResMsg resMsg={resMsg} />
      </form>
    </ProtectedRoute>
  );
}
