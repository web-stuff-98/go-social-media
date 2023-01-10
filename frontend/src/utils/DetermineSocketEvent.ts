import { IPrivateMessage } from "../components/layout/chat/Chat";

export type ChangeData = {
  TYPE: "CHANGE";
  METHOD: SocketEventChangeMethod;
  ENTITY: SocketEventChangeEntityType;
  DATA: object & { ID: string };
};

export type ResponseMessageData = {
  TYPE: "RESPONSE_MESSAGE";
  DATA: object & { ID: string };
};

export type PrivateMessageData = {
  TYPE: "PRIVATE_MESSAGE";
  DATA: object & IPrivateMessage;
};

export type SocketEventChangeMethod =
  | "UPDATE"
  | "INSERT"
  | "DELETE"
  | "UPDATE_IMAGE";
export type SocketEventChangeEntityType =
  | "ROOM"
  | "POST"
  | "POST_COMMENT"
  | "CHAT_MESSAGE"
  | "USER";

export function instanceOfChangeData(object: any): object is ChangeData {
  return object.TYPE === "CHANGE";
}
export function instanceOfResponseMessageData(
  object: any
): object is ResponseMessageData {
  return object.TYPE === "RESPONSE_MESSAGE";
}

export function instanceOfPrivateMessageData(
  object: any
): object is PrivateMessageData {
  return object.TYPE === "PRIVATE_MESSAGE";
}
