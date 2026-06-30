/**
 * Shared API envelope + error contract from docs/api-spec.md.
 * Field names are snake_case on the wire.
 */

export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
}

export interface DataResponse<T> {
  data: T;
}

export interface CollectionResponse<T> {
  data: T[];
  meta?: PaginationMeta;
}

export interface ApiErrorDetail {
  field: string;
  message: string;
}

export type ApiErrorCode =
  | "VALIDATION_ERROR"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "NOT_FOUND"
  | "CONFLICT"
  | "RATE_LIMITED"
  | "INTERNAL";

/**
 * Known error codes plus any future string the server may add. The `& {}`
 * preserves autocomplete for the known codes without collapsing to `string`.
 */
export type ApiErrorCodeValue = ApiErrorCode | (string & {});

export interface ApiErrorBody {
  error: {
    code: ApiErrorCodeValue;
    message: string;
    details?: ApiErrorDetail[];
  };
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
}
