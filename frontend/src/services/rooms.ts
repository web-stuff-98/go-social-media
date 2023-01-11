import { IRoom } from "../context/RoomsContext";
import { makeRequest } from "./makeRequest";

const createRoom = (data: Pick<IRoom, "name">) =>
  makeRequest("/api/rooms", {
    method: "POST",
    withCredentials: true,
    data,
  });

const updateRoom = (data: Pick<IRoom, "name" | "ID">) =>
  makeRequest(`/api/rooms/${data.ID}/update`, {
    method: "PATCH",
    withCredentials: true,
  });

const deleteRoom = (id: string) =>
  makeRequest(`/api/rooms/${id}/delete`, {
    withCredentials: true,
    method: "DELETE",
  });

export { createRoom, updateRoom, deleteRoom };
