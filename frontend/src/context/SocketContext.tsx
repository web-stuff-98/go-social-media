import { useEffect, useState, useContext, createContext } from "react";
import type { ReactNode } from "react";

/*
Change events (DELETE, INSERT, UPDATE) come through from the server like this :
{
  "TYPE": "CHANGE",
  "METHOD": SocketEventChangeMethod,
  "ENTITY": SocketEventEntityType,
  "DATA": "{ "ID":"ABCD" }", <- ID is included in data for deletes
}
*/

export const SocketContext = createContext<{
  socket?: WebSocket;
  connectSocket: () => void;
  reconnectSocket: () => void;
  openSubscription: (name: string) => void;
  closeSubscription: (name: string) => void;
  sendIfPossible: (data: string) => void;
}>({
  socket: undefined,
  connectSocket: () => {},
  reconnectSocket: () => {},
  openSubscription: () => {},
  closeSubscription: () => {},
  sendIfPossible: () => {},
});

export const SocketProvider = ({ children }: { children: ReactNode }) => {
  const [socket, setSocket] = useState<WebSocket | undefined>(undefined);

  // Store subscriptions in state so that if the websocket reconnects the subscriptions can be opened back up again
  const [openSubscriptions, setOpenSubscriptions] = useState<string[]>([]);
  // If a message couldn't be sent because the socket isn't properly connected, queue it and wait for the socket
  const [sendQueue, setSendQueue] = useState<string[]>([]);

  const queueSocketMessage = (msg: string) => setSendQueue((o) => [...o, msg]);

  const sendIfPossible = (data: string) => {
    if (socket && socket.readyState === 1) {
      socket.send(data);
    } else {
      queueSocketMessage(data);
    }
  };

  const onOpen = () => {
    if (openSubscriptions.length !== 0) {
      socket?.send(
        JSON.stringify({
          event_type: "OPEN_SUBSCRIPTIONS",
          names: openSubscriptions,
        })
      );
    }
    if (sendQueue.length !== 0) {
      sendQueue.forEach((msg) => {
        socket?.send(msg);
      });
      setSendQueue([]);
    }
  };

  const reconnectSocket = () => {
    if (!socket) return connectSocket();
    socket.close();
  };

  const connectSocket = () => {
    const socket = new WebSocket(
      process.env.NODE_ENV === "development" ||
      window.location.origin === "http://localhost:8080"
        ? "ws://localhost:8080/api/ws"
        : "wss://go-social-media-js.herokuapp.com/api/ws"
    );
    setSocket(socket);
  };

  const openSubscription = (subscriptionName: string) => {
    sendIfPossible(
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
    sendIfPossible(
      JSON.stringify({
        event_type: "CLOSE_SUBSCRIPTION",
        name: subscriptionName,
      })
    );
    setOpenSubscriptions((o) => [...o.filter((o) => o !== subscriptionName)]);
  };

  useEffect(() => {
    if (socket) {
      socket.addEventListener("open", onOpen);
      socket.addEventListener("close", connectSocket);
    } else connectSocket();
    return () => {
      if (socket) {
        socket.removeEventListener("open", onOpen);
        socket.removeEventListener("close", connectSocket);
        socket.close();
      }
    };
    // eslint-disable-next-line
  }, [socket]);

  return (
    <SocketContext.Provider
      value={{
        socket,
        connectSocket,
        openSubscription,
        closeSubscription,
        reconnectSocket,
        sendIfPossible,
      }}
    >
      {children}
    </SocketContext.Provider>
  );
};

const useSocket = () => useContext(SocketContext);
export default useSocket;
