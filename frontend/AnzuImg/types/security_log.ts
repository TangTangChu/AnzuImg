export interface SecurityLog {
    id: number;
    category: string;
    level: "info" | "warning" | "error";
    action: string;
    message: string;
    method?: string;
    path?: string;
    ip_address: string;
    username: string;
    created_at: string;
}

export interface SecurityLogListResponse {
    data: SecurityLog[];
    total: number;
    page: number;
    size: number;
}
