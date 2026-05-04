export type AppLogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL";

export interface AppLog {
    id: number;
    created_at: string;
    level: AppLogLevel | string;
    module: string;
    message: string;
    request_id?: string;
    ip_address?: string;
}

export interface AppLogListResponse {
    data: AppLog[];
    total: number;
    page: number;
    size: number;
}

export type LogSource = "app" | "security" | "token";

export interface LogFilter {
    search?: string;
    level?: string;
    module?: string;
    ip?: string;
    action?: string;
    start_date?: string;
    end_date?: string;
    failed_only?: boolean;
}
