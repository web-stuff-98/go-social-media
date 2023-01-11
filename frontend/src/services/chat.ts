import { makeRequest } from "./makeRequest";

const getConversations = () =>
  makeRequest("/api/account/conversations", {
    withCredentials: true,
  });

export { getConversations };
