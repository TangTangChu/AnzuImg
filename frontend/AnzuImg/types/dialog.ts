export enum DialogType {
    ALERT = 'alert',
    CONFIRM = 'confirm',
    PROMPT = 'prompt',
    CUSTOM = 'custom'
}

export enum DialogSize {
    SM = 'sm',
    MD = 'md',
    LG = 'lg',
    XL = 'xl',
    FULL = 'full'
}

export enum DialogPosition {
    CENTER = 'center',
    TOP = 'top',
    BOTTOM = 'bottom',
    LEFT = 'left',
    RIGHT = 'right',
    TOP_LEFT = 'top-left',
    TOP_RIGHT = 'top-right',
    BOTTOM_LEFT = 'bottom-left',
    BOTTOM_RIGHT = 'bottom-right'
}

export enum DialogVariant {
    DEFAULT = 'default',
    DESTRUCTIVE = 'destructive',
    SUCCESS = 'success',
    WARNING = 'warning',
    INFO = 'info'
}

export interface DialogAction {
    text: string
    handler?: () => void | Promise<void>
    primary?: boolean
    variant?: 'filled' | 'outlined' | 'text' | 'elevated' | 'tonal'
    disabled?: boolean
    loading?: boolean
    icon?: any
}

export interface DialogOptions {
    title?: string
    message?: string
    type?: DialogType
    size?: DialogSize
    position?: DialogPosition
    variant?: DialogVariant
    actions?: DialogAction[]
    showCloseButton?: boolean
    closeOnClickOutside?: boolean
    closeOnEsc?: boolean
    showOverlay?: boolean
    overlayClickable?: boolean
    persistent?: boolean
    maxWidth?: string
    minWidth?: string
    maxHeight?: string
    minHeight?: string
    customClass?: string
    icon?: any
    component?: any
    componentProps?: Record<string, any>
    resolve?: (value: any) => void
    reject?: (reason?: any) => void
}

export interface Dialog extends DialogOptions {
    id: number
    visible: boolean
}