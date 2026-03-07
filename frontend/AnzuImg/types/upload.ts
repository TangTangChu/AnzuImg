export interface UploadFileItem {
  file: File;
  previewUrl: string;
  description: string;
  tags: string[];
  routes: string[];
  customName: string;
  status: "pending" | "success" | "error";
  error?: string;
  resultUrl?: string;
  client_index?: number;
}

export interface UploadResultItem {
  client_index?: number;
  success: boolean;
  file_name?: string;
  url?: string;
  code?: string;
  message?: string;
}
