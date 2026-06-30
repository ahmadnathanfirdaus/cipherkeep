import { AxiosError } from "axios";
import type { ApiErrorBody, ApiErrorDetail } from "@/types";

export interface ParsedApiError {
  code: string;
  message: string;
  details: ApiErrorDetail[];
  status?: number;
}

const isApiErrorBody = (value: unknown): value is ApiErrorBody => {
  if (typeof value !== "object" || value === null) return false;
  const candidate = value as Record<string, unknown>;
  const err = candidate.error;
  if (typeof err !== "object" || err === null) return false;
  const errObj = err as Record<string, unknown>;
  return (
    typeof errObj.code === "string" && typeof errObj.message === "string"
  );
};

/** Normalize any thrown value into a displayable API error. */
export const parseApiError = (error: unknown): ParsedApiError => {
  if (error instanceof AxiosError) {
    const body: unknown = error.response?.data;
    if (isApiErrorBody(body)) {
      return {
        code: body.error.code,
        message: body.error.message,
        details: body.error.details ?? [],
        status: error.response?.status,
      };
    }
    if (error.code === "ERR_NETWORK") {
      return {
        code: "NETWORK_ERROR",
        message: "Cannot reach the server. Check your connection.",
        details: [],
      };
    }
    return {
      code: "UNKNOWN",
      message: error.message,
      details: [],
      status: error.response?.status,
    };
  }

  if (error instanceof Error) {
    return { code: "UNKNOWN", message: error.message, details: [] };
  }

  return {
    code: "UNKNOWN",
    message: "An unexpected error occurred.",
    details: [],
  };
};

export const getApiErrorMessage = (error: unknown): string =>
  parseApiError(error).message;
