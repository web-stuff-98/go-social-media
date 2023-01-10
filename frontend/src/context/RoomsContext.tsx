import { useState, useContext, createContext, useCallback } from "react";
import type { ReactNode } from "react";
import { makeRequest } from "../services/makeRequest";

export interface IRoom {
  ID: string;
  name: string;
  author_id: string;
  img_blur?: string;
  img_url?: string;
}

const RoomsContext = createContext<{
  rooms: IRoom[];
  updateRoomData: (
    data: Omit<Partial<IRoom>, "event_type">,
    dontAddIfNoId?: boolean
  ) => void;
  deleteRoom: (id: string) => Promise<void>;
  setAllRooms: (rooms: IRoom[]) => void;
  ownRooms: boolean;
  setOwnRooms: (to: boolean) => void;
  getRoomData: (id: string) => IRoom | undefined;
  deleteRoomsByAuthor: (uid: string) => void;
}>({
  rooms: [],
  updateRoomData: () => {},
  deleteRoom: () => new Promise(() => {}),
  setAllRooms: () => {},
  ownRooms: false,
  setOwnRooms: () => {},
  getRoomData: () => undefined,
  deleteRoomsByAuthor: () => {},
});

export const RoomsProvider = ({ children }: { children: ReactNode }) => {
  const [rooms, setRooms] = useState<IRoom[]>([]);
  const [ownRooms, setOwnRooms] = useState(false);

  const updateRoomData = (data: Partial<IRoom>, dontAddIfNoId?: boolean) => {
    setRooms((old) => {
      let newRooms = old;
      const i = old.findIndex((r) => r.ID === data.ID);
      if (i !== -1) {
        newRooms[i] = { ...newRooms[i], ...(data as Omit<any, "img_url">) };
      } else if (!dontAddIfNoId) {
        newRooms = [...newRooms, data as Omit<IRoom, "img_url">];
      }
      return [...newRooms];
    });
  };

  const getRoomData = useCallback(
    (id: string) => rooms.find((r) => r.ID === id),
    [rooms]
  );

  const deleteRoomsByAuthor = (uid: string) =>
    setRooms((o) => [...o.filter((r) => r.author_id !== uid)]);

  const setAllRooms = (rooms: IRoom[]) => setRooms(rooms);

  const deleteRoom = (id: string) => {
    return new Promise<void>((resolve, reject) =>
      makeRequest(`/api/room/${id}`, {
        method: "DELETE",
        withCredentials: true,
      })
        .then(() => {
          setRooms((old) => [...old.filter((r) => r.ID !== id)]);
          resolve();
        })
        .catch((e) => reject(e))
    );
  };

  return (
    <RoomsContext.Provider
      value={{
        rooms,
        updateRoomData,
        setAllRooms,
        setOwnRooms,
        ownRooms,
        deleteRoom,
        getRoomData,
        deleteRoomsByAuthor,
      }}
    >
      {children}
    </RoomsContext.Provider>
  );
};

export const useRooms = () => useContext(RoomsContext);
