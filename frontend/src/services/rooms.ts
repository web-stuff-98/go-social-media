import axios from "axios";
import { IRoomCard } from "../interfaces/ChatInterfaces";
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

const getRoomPrivateData = (id: string) =>
  makeRequest(`/api/rooms/${id}/private-data`, {
    withCredentials: true,
  });

const getRoomImage = async (id: string) => {
  const data = await makeRequest(`/api/rooms/${id}/image`, {
    responseType: "arraybuffer",
  });
  const blob = new Blob([data], { type: "image/jpeg" });
  return URL.createObjectURL(blob);
};

const getRoomImageAsBlob = async (id: string) => {
  const data = await makeRequest(`/api/rooms/${id}/image`, {
    responseType: "arraybuffer",
  });
  return new Blob([data], { type: "image/jpeg" });
};

const getRandomRoomImage = async () => {
  const res = await axios({
    url: "https://picsum.photos/500/200",
    responseType: "arraybuffer",
  });
  const file = new File([res.data], "image.jpg", { type: "image/jpeg" });
  return file;
};

const deleteRoom = (id: string) =>
  makeRequest(`/api/rooms/${id}/delete`, {
    withCredentials: true,
    method: "DELETE",
  });

const getRoomPage = (page: number, term: string, own: boolean) =>
  makeRequest(
    `/api/rooms/page/${page}${
      term ? `?term=${term.replaceAll(" ", "+")}` : ""
    }${own ? `${term ? "&" : "?"}own=true` : ""}`,
    {
      withCredentials: true,
    }
  );

const getOwnRooms = () =>
  makeRequest("/api/rooms/own", { withCredentials: true });

const uploadRoomImage = (file: File, id: string) => {
  const data = new FormData();
  data.append("file", file);
  return makeRequest(`/api/rooms/${id}/image`, {
    method: "POST",
    withCredentials: true,
    data,
  });
};

const inviteToRoom = (uid: string, roomId: string) =>
  makeRequest(`/api/rooms/${roomId}/invite?uid=${uid}`, {
    withCredentials: true,
    method: "POST",
  });

const banFromRoom = (uid: string, roomId: string) =>
  makeRequest(`/api/rooms/${roomId}/ban?uid=${uid}`, {
    withCredentials: true,
    method: "POST",
  });

const unbanFromRoom = (uid: string, roomId: string) =>
  makeRequest(`/api/rooms/${roomId}/unban?uid=${uid}`, {
    withCredentials: true,
    method: "POST",
  });

const declineInvite = (uid: string, msgId: string, roomId: string) =>
  makeRequest(`/api/rooms/${roomId}/invite/decline/${msgId}?uid=${uid}`, {
    withCredentials: true,
    method: "POST",
  });

const acceptInvite = (uid: string, msgId: string, roomId: string) =>
  makeRequest(`/api/rooms/${roomId}/invite/accept/${msgId}?uid=${uid}`, {
    withCredentials: true,
    method: "POST",
  });

export {
  createRoom,
  getRoomImage,
  updateRoom,
  uploadRoomImage,
  getRoomPrivateData,
  deleteRoom,
  getRoomPage,
  getRoom,
  getRoomImageAsBlob,
  getRandomRoomImage,
  getOwnRooms,
  inviteToRoom,
  banFromRoom,
  unbanFromRoom,
  declineInvite,
  acceptInvite,
};
