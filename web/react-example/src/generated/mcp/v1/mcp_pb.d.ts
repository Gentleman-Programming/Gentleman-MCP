import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb'; // proto import: "google/protobuf/timestamp.proto"


export class RegisterRequest extends jspb.Message {
  getTenantId(): string;
  setTenantId(value: string): RegisterRequest;

  getAgentId(): string;
  setAgentId(value: string): RegisterRequest;

  getModel(): string;
  setModel(value: string): RegisterRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterRequest): RegisterRequest.AsObject;
  static serializeBinaryToWriter(message: RegisterRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterRequest;
  static deserializeBinaryFromReader(message: RegisterRequest, reader: jspb.BinaryReader): RegisterRequest;
}

export namespace RegisterRequest {
  export type AsObject = {
    tenantId: string,
    agentId: string,
    model: string,
  }
}

export class RegisterResponse extends jspb.Message {
  getSessionId(): string;
  setSessionId(value: string): RegisterResponse;

  getJwtToken(): string;
  setJwtToken(value: string): RegisterResponse;

  getExpiresAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setExpiresAt(value?: google_protobuf_timestamp_pb.Timestamp): RegisterResponse;
  hasExpiresAt(): boolean;
  clearExpiresAt(): RegisterResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterResponse): RegisterResponse.AsObject;
  static serializeBinaryToWriter(message: RegisterResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterResponse;
  static deserializeBinaryFromReader(message: RegisterResponse, reader: jspb.BinaryReader): RegisterResponse;
}

export namespace RegisterResponse {
  export type AsObject = {
    sessionId: string,
    jwtToken: string,
    expiresAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class AuthRequest extends jspb.Message {
  getJwtToken(): string;
  setJwtToken(value: string): AuthRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AuthRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AuthRequest): AuthRequest.AsObject;
  static serializeBinaryToWriter(message: AuthRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AuthRequest;
  static deserializeBinaryFromReader(message: AuthRequest, reader: jspb.BinaryReader): AuthRequest;
}

export namespace AuthRequest {
  export type AsObject = {
    jwtToken: string,
  }
}

export class AuthResponse extends jspb.Message {
  getValid(): boolean;
  setValid(value: boolean): AuthResponse;

  getTenantId(): string;
  setTenantId(value: string): AuthResponse;

  getAgentId(): string;
  setAgentId(value: string): AuthResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AuthResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AuthResponse): AuthResponse.AsObject;
  static serializeBinaryToWriter(message: AuthResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AuthResponse;
  static deserializeBinaryFromReader(message: AuthResponse, reader: jspb.BinaryReader): AuthResponse;
}

export namespace AuthResponse {
  export type AsObject = {
    valid: boolean,
    tenantId: string,
    agentId: string,
  }
}

export class ChatMessage extends jspb.Message {
  getMessageId(): string;
  setMessageId(value: string): ChatMessage;

  getSessionId(): string;
  setSessionId(value: string): ChatMessage;

  getContent(): string;
  setContent(value: string): ChatMessage;

  getType(): MessageType;
  setType(value: MessageType): ChatMessage;

  getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): ChatMessage;
  hasTimestamp(): boolean;
  clearTimestamp(): ChatMessage;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatMessage.AsObject;
  static toObject(includeInstance: boolean, msg: ChatMessage): ChatMessage.AsObject;
  static serializeBinaryToWriter(message: ChatMessage, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatMessage;
  static deserializeBinaryFromReader(message: ChatMessage, reader: jspb.BinaryReader): ChatMessage;
}

export namespace ChatMessage {
  export type AsObject = {
    messageId: string,
    sessionId: string,
    content: string,
    type: MessageType,
    timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export class SingleChatRequest extends jspb.Message {
  getSessionId(): string;
  setSessionId(value: string): SingleChatRequest;

  getContent(): string;
  setContent(value: string): SingleChatRequest;

  getModel(): string;
  setModel(value: string): SingleChatRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SingleChatRequest): SingleChatRequest.AsObject;
  static serializeBinaryToWriter(message: SingleChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleChatRequest;
  static deserializeBinaryFromReader(message: SingleChatRequest, reader: jspb.BinaryReader): SingleChatRequest;
}

export namespace SingleChatRequest {
  export type AsObject = {
    sessionId: string,
    content: string,
    model: string,
  }
}

export class SingleChatResponse extends jspb.Message {
  getContent(): string;
  setContent(value: string): SingleChatResponse;

  getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): SingleChatResponse;
  hasTimestamp(): boolean;
  clearTimestamp(): SingleChatResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SingleChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SingleChatResponse): SingleChatResponse.AsObject;
  static serializeBinaryToWriter(message: SingleChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SingleChatResponse;
  static deserializeBinaryFromReader(message: SingleChatResponse, reader: jspb.BinaryReader): SingleChatResponse;
}

export namespace SingleChatResponse {
  export type AsObject = {
    content: string,
    timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
  }
}

export enum MessageType { 
  MESSAGE_TYPE_UNSPECIFIED = 0,
  MESSAGE_TYPE_USER = 1,
  MESSAGE_TYPE_ASSISTANT = 2,
  MESSAGE_TYPE_SYSTEM = 3,
}
