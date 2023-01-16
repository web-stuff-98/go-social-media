import {
  useEffect,
  useState,
  useContext,
  createContext,
  useCallback,
} from "react";
import type { ReactNode } from "react";
import { useModal } from "./ModalContext";
import { instanceOfResponseMessageData } from "../utils/DetermineSocketEvent";

/*
Change events (DELETE, INSERT, UPDATE) come through from the server like this :
{
  "TYPE": "CHANGE",
  "METHOD": SocketEventChangeMethod,
  "ENTITY": SocketEventEntityType,
  "DATA": "{ "ID":"ABCD" }", <- ID is included in data for deletes
}
*/

const SocketContext = createContext<{
  socket?: WebSocket;
  connectSocket: () => void;
  reconnectSocket: () => void;
  openSubscription: (name: string) => void;
  closeSubscription: (name: string) => void;
}>({
  socket: undefined,
  connectSocket: () => {},
  reconnectSocket: () => {},
  openSubscription: () => {},
  closeSubscription: () => {},
});

export const SocketProvider = ({ children }: { children: ReactNode }) => {
  const { openModal } = useModal();

  const [socket, setSocket] = useState<WebSocket | undefined>(undefined);

  // Store subscriptions in state so that if the websocket reconnects the subscriptions can be opened back up again
  const [openSubscriptions, setOpenSubscriptions] = useState<string[]>([]);

  const onOpen = () => {
    if (openSubscriptions.length !== 0)
      socket?.send(
        JSON.stringify({
          event_type: "OPEN_SUBSCRIPTIONS",
          names: openSubscriptions,
        })
      );
  };

  const reconnectSocket = () => {
    if (!socket) return connectSocket();
    socket.close();
  };

  const connectSocket = () => {
    const socket = new WebSocket(
      process.env.NODE_ENV === "development"
        ? "ws://localhost:8080/api/ws"
        : "wss://go-social-media-js.herokuapp.com/api/ws"
    );
    setSocket(socket);
  };

  const openSubscription = (subscriptionName: string) => {
    socket?.send(
      JSON.stringify({
        event_type: "OPEN_SUBSCRIPTION",
        name: subscriptionName,
      })
    );
    if (!subscriptionName.includes("inbox="))
      setOpenSubscriptions((o) => [
        ...o.filter((o) => o !== subscriptionName),
        subscriptionName,
      ]);
  };
  const closeSubscription = (subscriptionName: string) => {
    socket?.send(
      JSON.stringify({
        event_type: "CLOSE_SUBSCRIPTION",
        name: subscriptionName,
      })
    );
    setOpenSubscriptions((o) => [...o.filter((o) => o !== subscriptionName)]);
  };

  const handleMessage = useCallback((e: MessageEvent) => {
    console.log(`Data before parse : `, e.data);
    const data = JSON.parse(e.data);
    console.log(data);
    data["DATA"] = JSON.parse(data["DATA"]);
    if (instanceOfResponseMessageData(e.data)) {
      openModal("Message", {
        msg: data.msg,
        err: data.err,
        pen: false,
      });
    }
  }, []);

  useEffect(() => {
    if (socket) {
      socket.addEventListener("open", onOpen);
      socket.addEventListener("close", connectSocket);
      socket.addEventListener("message", handleMessage);
    } else connectSocket();
    return () => {
      if (socket) {
        socket.removeEventListener("open", onOpen);
        socket.removeEventListener("close", connectSocket);
        socket.removeEventListener("message", handleMessage);
        socket.close();
      }
    };
  }, [socket]);

  return (
    <SocketContext.Provider
      value={{
        socket,
        connectSocket,
        openSubscription,
        closeSubscription,
        reconnectSocket,
      }}
    >
      {children}
    </SocketContext.Provider>
  );
};

const useSocket = () => useContext(SocketContext);
export default useSocket;
