import { Navigate } from "react-router-dom";
import { IUser } from "../interfaces/GeneralInterfaces";

export default function ProtectedRoute<ReactNode>({
  user,
  children,
}: {
  user?: IUser;
  children: ReactNode;
}) {
  if (!user) {
    return <Navigate to={"/login"} />;
  } else {
    return children;
  }
}
