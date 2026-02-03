export interface Image {
    hash: string;
    file_name: string;
    size?: number;
    width?: number;
    height?: number;
    created_at?: string;
    tags?: string[];
    description?: string;
    mime?: string;
}

export interface ImageDetail extends Image {
    mime_type?: string;
    updated_at?: string;
    routes?: string[];
}

export interface ImageListResponse {
    data: Image[];
    total: number;
    page: number;
    size: number;
}

export interface TagSummary {
    tag: string;
    count: number;
}

export interface TagListResponse {
    data: TagSummary[];
}

export interface ImageModalProps {
    image: Image | null;
    visible: boolean;
    currentIndex: number;
    totalImages: number;
    hasPrevious: boolean;
    hasNext: boolean;
}

export interface ImageModalEmits {
    (e: 'update:visible', value: boolean): void;
    (e: 'close'): void;
    (e: 'previous'): void;
    (e: 'next'): void;
    (e: 'copy-link'): void;
    (e: 'download'): void;
    (e: 'delete', hash: string): void;
}
