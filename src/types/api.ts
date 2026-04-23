export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T | null;
}

export interface ApiError {
  code: number;
  message: string;
  data?: unknown;
}

export const ErrorCodes = {
  Success: 0,
  InvalidParam: 1001,
  NotFound: 1002,
  DuplicateResource: 1003,
  DockerConnFailed: 2001,
  BuildFailed: 2002,
  DeployFailed: 2003,
  ImageNotFound: 2004,
  ContainerNotRunning: 2005,
  RollbackFailed: 2006,
  InternalError: 3001,
  DatabaseError: 3002,
  ConfigError: 3003,
  RegistryAuthFailed: 4001,
  RegistryConnFailed: 4002,
  NetworkError: 4003,
} as const;

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes];

export class FlowCIError extends Error {
  constructor(
    public code: number,
    message: string,
    public data?: unknown
  ) {
    super(message);
    this.name = 'FlowCIError';
  }
}

export class RequestError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(1001, message, data);
    this.name = 'RequestError';
  }
}

export class BuildError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(2002, message, data);
    this.name = 'BuildError';
  }
}

export class DeployError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(2003, message, data);
    this.name = 'DeployError';
  }
}
