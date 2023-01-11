import { makeRequest } from "./makeRequest";

const getConversations = () =>
  makeRequest("/api/account/conversations", {
    withCredentials: true,
  });

const getConversation = (uid: string) =>
  makeRequest(`/api/account/conversation/${uid}`, { withCredentials: true });

export { getConversations, getConversation };
