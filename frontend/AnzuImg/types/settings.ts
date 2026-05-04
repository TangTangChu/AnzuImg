export type FieldType =
    | "string"
    | "multiline"
    | "int"
    | "int64"
    | "bool"
    | "enum"
    | "list";

export type FieldGroup =
    | "uploads"
    | "session"
    | "login_security"
    | "password_policy"
    | "network"
    | "logs"
    | "stepup";

export interface FieldSchema {
    key: string;
    group: FieldGroup;
    type: FieldType;
    default: unknown;
    min?: number;
    max?: number;
    options?: string[];
    sensitive?: boolean;
    requires_restart?: boolean;
}

export interface FieldValue {
    key: string;
    value: unknown;
    overridden_in_db: boolean;
}

export interface SettingsResponse {
    schema: FieldSchema[];
    values: FieldValue[];
    allow_web_modify: boolean;
    bootstrap_notice?: string;
}
