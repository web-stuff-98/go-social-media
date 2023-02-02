import { useState, useContext, createContext, useEffect } from "react";
import type { ReactNode } from "react";
import { makeRequest } from "../services/makeRequest";
import useSocket from "./SocketContext";
import { useModal } from "./ModalContext";
import { IUser } from "../interfaces/GeneralInterfaces";

export const AuthContext = createContext<{
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
  const { socket, reconnectSocket, openSubscription } = useSocket();
  const { openModal } = useModal();

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

  const logout = async () => {
    try {
      await makeRequest("/api/account/logout", {
        method: "POST",
        withCredentials: true,
      });
      setUser(undefined);
      reconnectSocket();
    } catch (e) {
      openModal("Message", {
        msg: `${e}`,
        err: true,
        pen: false,
      });
    }
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
      .then((data: any) => {
        setUser(data.ID ? data : undefined);
      })
      .catch((e: unknown) => {
        setUser(undefined);
        reconnectSocket();
      });
  }, []);

  useEffect(() => {
    const i = setInterval(async () => {
      try {
        const data = await makeRequest("/api/account/refresh", {
          withCredentials: true,
          method: "POST",
        });
        if (!data.ID) setUser(undefined);
      } catch (e) {
        setUser(undefined);
      }
      //Refresh token every 90 seconds. Token expires after 120 seconds.
    }, 90000);
    return () => {
      clearInterval(i);
    };
  }, [user]);

  const handleSocketOnOpen = () => {
    if (user) {
      openSubscription(`inbox=${user.ID}`);
      openSubscription(`notifications=${user.ID}`);
    }
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
