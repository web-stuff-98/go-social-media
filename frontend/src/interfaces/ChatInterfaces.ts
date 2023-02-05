export interface IMsgAttachmentProgress {
  ratio: number;
  failed: boolean;
  pending: boolean;
}

export interface IAttachmentData {
  name: string;
  type: string;
  size: number;
  length: number;
}

export interface IPrivateMessage {
  ID: string;
  uid: string;
  content: string;
  created_at: string;
  updated_at: string;
  has_attachment: boolean;
  recipient_id: string;
  attachment_progress?: IMsgAttachmentProgress;
  attachment_metadata?: IAttachmentData;
  invitation?: boolean;
  invitation_accepted?: boolean;
  invitation_declined?: boolean;
}

export type IConversation = {
  uid: string;
  messages: IPrivateMessage[];
};

export interface IRoomMessage {
  ID: string;
  uid: string;
  content: string;
  created_at: string;
  updated_at: string;
  has_attachment: boolean;
  attachment_progress?: IMsgAttachmentProgress;
  attachment_metadata?: IAttachmentData;
}

export interface IRoom extends IRoomCard {
  messages: IRoomMessage[];
}

export interface IRoomCard {
  ID: string;
  name: string;
  author_id: string;
  img_blur?: string;
  img_url?: string;
  private?: boolean;
}
