export interface IResMsg {
  msg: string;
  err: boolean;
  pen: boolean;
}

export interface IUser {
  ID: string;
  username: string;
  base64pfp?: string;
}

export interface IDimensions {
  width: number;
  height: number;
}

export interface IPosition {
  top: number;
  left: number;
}
