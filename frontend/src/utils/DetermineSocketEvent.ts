export type ChangeData = {
  TYPE: "CHANGE";
  METHOD: SocketEventChangeMethod;
  ENTITY: SocketEventChangeEntityType;
  DATA: object & { ID: string };
};

export type MessageData = {
  TYPE: "MESSAGE";
  DATA: object & { ID: string };
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
export function instanceOfMessageData(object: any): object is MessageData {
  return object.TYPE === "MESSAGE";
}
