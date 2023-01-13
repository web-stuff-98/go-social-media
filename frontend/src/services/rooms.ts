import { IRoomCard } from "../components/layout/chat/Rooms";
import { makeRequest } from "./makeRequest";

const createRoom = (data: Pick<IRoomCard, "name">) =>
  makeRequest("/api/rooms", {
    method: "POST",
    withCredentials: true,
    data,
  });

const updateRoom = (data: Pick<IRoomCard, "name" | "ID">) =>
  makeRequest(`/api/rooms/${data.ID}/update`, {
    method: "PATCH",
    withCredentials: true,
  });

const getRoom = (id: string) =>
  makeRequest(`/api/rooms/${id}`, {
    withCredentials: true,
  });

const getRoomImage = async (id: string) => {
  const data = await makeRequest(`/api/rooms/${id}/image`, {
    responseType: "arraybuffer",
  });
  const blob = new Blob([data], { type: "image/jpeg" });
  return URL.createObjectURL(blob);
};

const deleteRoom = (id: string) =>
  makeRequest(`/api/rooms/${id}/delete`, {
    withCredentials: true,
    method: "DELETE",
  });

const getRoomPage = (page: number, term: string) =>
  makeRequest(`/api/rooms/page/${page}${term ? `?term=${term.replaceAll(" ", "+")}` : ""}`, {
    withCredentials: true,
  });

const uploadRoomImage = (file: File, id: string) => {
  const data = new FormData();
  data.append("file", file);
  return makeRequest(`/api/rooms/${id}/image`, {
    method: "POST",
    withCredentials: true,
    data,
  });
};

export {
  createRoom,
  getRoomImage,
  updateRoom,
  uploadRoomImage,
  deleteRoom,
  getRoomPage,
  getRoom,
};
