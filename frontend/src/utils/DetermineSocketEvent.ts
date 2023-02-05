import { IPrivateMessage, IRoomMessage } from "../interfaces/ChatInterfaces";

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
export type PrivateMessageDeleteData = {
  TYPE: "PRIVATE_MESSAGE_DELETE";
  DATA: {
    ID: string;
    recipient_id: string;
  };
};
export type PrivateMessageUpdateData = {
  TYPE: "PRIVATE_MESSAGE_UPDATE";
  DATA: {
    ID: string;
    content: string;
    recipient_id: string;
  };
};
export type RoomMessageData = {
  TYPE: "ROOM_MESSAGE";
  DATA: IRoomMessage;
};
export type RoomMessageDeleteData = {
  TYPE: "ROOM_MESSAGE_DELETE";
  DATA: {
    ID: string;
  };
};
export type RoomMessageUpdateData = {
  TYPE: "ROOM_MESSAGE_UPDATE";
  DATA: {
    ID: string;
    content: string;
  };
};
export type PostVoteData = {
  TYPE: "POST_VOTE";
  DATA: { ID: string; is_upvote: boolean; remove: boolean };
};
export type PostCommentVoteData = {
  TYPE: "POST_COMMENT_VOTE";
  DATA: { ID: string; is_upvote: boolean; remove: boolean };
};
export type AttachmentProgressData = {
  TYPE: "ATTACHMENT_PROGRESS";
  DATA: { ID: string; failed: boolean; pending: boolean; ratio: number };
};
export type AttachmentCompleteData = {
  TYPE: "ATTACHMENT_COMPLETE";
  DATA: {
    ID: string;
    type: string;
    size: number;
    name: string;
    length: number;
  };
};
export type NotificationsData = {
  TYPE: "NOTIFICATIONS";
  DATA: string;
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
  | "MEMBER"
  | "BANNED"
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

export function instanceOfPrivateMessageDeleteData(
  object: any
): object is PrivateMessageDeleteData {
  return object.TYPE === "PRIVATE_MESSAGE_DELETE";
}

export function instanceOfPrivateMessageUpdateData(
  object: any
): object is PrivateMessageUpdateData {
  return object.TYPE === "PRIVATE_MESSAGE_UPDATE";
}

export function instanceOfRoomMessageData(
  object: any
): object is RoomMessageData {
  return object.TYPE === "ROOM_MESSAGE";
}

export function instanceOfRoomMessageDeleteData(
  object: any
): object is RoomMessageDeleteData {
  return object.TYPE === "ROOM_MESSAGE_DELETE";
}

export function instanceOfRoomMessageUpdateData(
  object: any
): object is RoomMessageUpdateData {
  return object.TYPE === "ROOM_MESSAGE_UPDATE";
}

export function instanceOfPostVoteData(object: any): object is PostVoteData {
  return object.TYPE === "POST_VOTE";
}

export function instanceOfPostCommentVoteData(
  object: any
): object is PostCommentVoteData {
  return object.TYPE === "POST_COMMENT_VOTE";
}

export function instanceOfAttachmentProgressData(
  object: any
): object is AttachmentProgressData {
  return object.TYPE === "ATTACHMENT_PROGRESS";
}

export function instanceOfAttachmentCompleteData(
  object: any
): object is AttachmentCompleteData {
  return object.TYPE === "ATTACHMENT_COMPLETE";
}

export function instanceOfNotificationsData(
  object: any
): object is NotificationsData {
  return object.TYPE === "NOTIFICATIONS";
}

/*
  Determine video chat events
*/

export type VidReceivingReturnedSignal = {
  TYPE: "VID_RECEIVING_RETURNED_SIGNAL";
  signal_json: string;
  uid: string;
};
export type VidAllUsersData = {
  TYPE: "VID_ALL_USERS";
  uids: string[];
};
export type VidUserLeftData = {
  TYPE: "VID_USER_LEFT";
  uid: string;
};
export type VidUserJoinedData = {
  TYPE: "VID_USER_JOINED";
  signal_json: string;
  caller_uid: string;
};

export function instanceOfReceivingReturnedSignal(
  object: any
): object is VidReceivingReturnedSignal {
  return object.TYPE === "VID_RECEIVING_RETURNED_SIGNAL";
}

export function instanceOfVidAllUsers(object: any): object is VidAllUsersData {
  return object.TYPE === "VID_ALL_USERS";
}

export function instanceOfVidUserLeft(object: any): object is VidUserLeftData {
  return object.TYPE === "VID_USER_LEFT";
}

export function instanceOfVidUserJoined(
  object: any
): object is VidUserJoinedData {
  return object.TYPE === "VID_USER_JOINED";
}
