export interface APIToken {
    id: number;
    name: string;
    token_type: string;
    ip_allowlist: string[];
    last_used_at: string | null;
    last_used_ip: string;
    created_at: string;
}

export interface CreateTokenResponse {
    token: APIToken;
    raw_token: string;
}

export interface APITokenLog {
    id: number;
    token_id: number;
    token_name: string;
    token_type: string;
    action: string;
    method: string;
    path: string;
    ip_address: string;
    user_agent: string;
    image_hash?: string;
    created_at: string;
}

export interface APITokenLogListResponse {
    data: APITokenLog[];
    total: number;
    page: number;
    size: number;
}
