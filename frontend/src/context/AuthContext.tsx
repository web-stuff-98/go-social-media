import {
  useState,
  useContext,
  createContext,
  useEffect,
} from "react";
import type { ReactNode } from "react";
import { makeRequest } from "../services/makeRequest";
import useSocket from "./SocketContext";

export interface IUser {
  ID: string;
  username: string;
  base64pfp?: string;
}

const AuthContext = createContext<{
  user?: IUser;
  login: (username: string, password: string) => void;
  logout: () => void;
  deleteAccount: () => void;
  register: (username: string, password: string) => void;
  updateUserState: (user: Partial<IUser>) => void;
}>({
  user: undefined,
  login: () => {},
  register: () => {},
  logout: () => {},
  deleteAccount: () => {},
  updateUserState: () => {},
});

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<IUser>();
  const { socket, reconnectSocket, openSubscription } =
    useSocket();

  const login = async (username: string, password: string) => {
    const user = await makeRequest("/api/account/login", {
      method: "POST",
      headers: { "Content-Type": "application/json;charset=UTF-8" },
      data: { username, password },
      withCredentials: true,
    });
    reconnectSocket();
    setUser(user);
  };

  const register = async (username: string, password: string) => {
    const user = await makeRequest("/api/account/register", {
      method: "POST",
      headers: { "Content-Type": "application/json;charset=UTF-8" },
      data: { username, password },
      withCredentials: true,
    });
    setUser(user);
    reconnectSocket();
  };

  const logout = () => {
    makeRequest("/api/account/logout", {
      method: "POST",
      withCredentials: true,
    }).finally(() => {
      setUser(undefined);
      reconnectSocket();
    });
  };

  const deleteAccount = async () => {
    await makeRequest("/api/account/delete", {
      withCredentials: true,
      method: "POST",
    });
    setUser(undefined);
  };

  useEffect(() => {
    makeRequest("/api/account/refresh", {
      withCredentials: true,
      method: "POST",
    })
      .then((data) => {
        setUser(data.ID ? data : undefined);
        reconnectSocket();
      })
      .catch((e) => {
        setUser(undefined);
        console.warn(e);
      });
  }, []);
  
  useEffect(() => {
    const i = setInterval(async () => {
      try {
        const data = await makeRequest("/api/account/refresh", {
          withCredentials: true,
          method: "POST",
        });
        if(!data.ID) setUser(undefined)
      } catch (e) {
        setUser(undefined);
        console.warn(e);
      }
      //Refresh token every 90 seconds. Token expires after 120 seconds.
    }, 90000);
    return () => {
      clearInterval(i);
    };
  }, [user]);

  const handleSocketOnOpen = () => {
    if (user) openSubscription(`inbox=${user.ID}`);
  };

  useEffect(() => {
    socket?.addEventListener("open", handleSocketOnOpen);
    return () => {
      socket?.removeEventListener("open", handleSocketOnOpen);
    };
  }, [socket, user]);

  const updateUserState = (user: Partial<IUser>) =>
    setUser((old) => ({ ...old, ...user } as IUser));

  return (
    <AuthContext.Provider
      value={{ user, login, register, logout, deleteAccount, updateUserState }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
