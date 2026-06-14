export type UploadSource = "file" | "url-server";

export interface UploadFileItem {
  file: File | null;
  previewUrl: string;
  description: string;
  tags: string[];
  routes: string[];
  customName: string;
  status: "pending" | "success" | "error";
  error?: string;
  resultUrl?: string;
  client_index?: number;
  source: UploadSource;
  sourceUrl?: string;
  displayName: string;
  displaySize: number;
  displayMime: string;
  displayWidth?: number;
  displayHeight?: number;
}

export interface UploadResultItem {
  client_index?: number;
  success: boolean;
  file_name?: string;
  url?: string;
  code?: string;
  message?: string;
}

export interface UrlSourceMetadata {
  url: string;
  client_index: number;
  description: string;
  tags: string[];
  routes: string[];
  custom_name: string;
}
