import { IPrivateMessage } from "../components/layout/chat/Conversations";
import { IRoomMessage } from "../components/layout/chat/Room";

export type ChangeData = {
  TYPE: "CHANGE";
  METHOD: SocketEventChangeMethod;
  ENTITY: SocketEventChangeEntityType;
  DATA: { ID: string };
};

export type ResponseMessageData = {
  TYPE: "RESPONSE_MESSAGE";
  DATA: { ID: string };
};

export type PrivateMessageData = {
  TYPE: "PRIVATE_MESSAGE";
  DATA: IPrivateMessage;
};

export type RoomMessageData = {
  TYPE: "ROOM_MESSAGE";
  DATA: IRoomMessage;
};

export type PostVoteData = {
  TYPE: "POST_VOTE";
  DATA: { ID: string; is_upvote: boolean; remove: boolean };
};

export type PostCommentVoteData = {
  TYPE: "POST_COMMENT_VOTE";
  DATA: { ID: string; is_upvote: boolean; remove: boolean };
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

export function instanceOfRoomMessageData(
  object: any
): object is RoomMessageData {
  return object.TYPE === "ROOM_MESSAGE";
}

export function instanceOfPostVoteData(object: any): object is PostVoteData {
  return object.TYPE === "POST_VOTE";
}

export function instanceOfPostCommentVoteData(
  object: any
): object is PostCommentVoteData {
  return object.TYPE === "POST_COMMENT_VOTE";
}
