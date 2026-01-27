export interface APIToken {
    id: number;
    name: string;
    ip_allowlist: string[];
    last_used_at: string | null;
    last_used_ip: string;
    created_at: string;
}

export interface CreateTokenResponse {
    token: APIToken;
    raw_token: string;
}
