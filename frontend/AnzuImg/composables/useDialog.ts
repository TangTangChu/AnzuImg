import { ref, markRaw } from "vue";
import {
    DialogType,
    type Dialog,
    type DialogOptions,
    type DialogSize,
    type DialogPosition,
    type DialogVariant,
    type DialogAction,
} from "~/types/dialog";

interface DialogInstance extends Dialog {
    resolve?: (value: any) => void;
    reject?: (reason?: any) => void;
}

export const enum DialogDismissCode {
    CLOSED = "dialog_closed",
    ALL_CLOSED = "all_dialogs_closed",
}

export class DialogDismissedError extends Error {
    code: DialogDismissCode;

    constructor(code: DialogDismissCode) {
        super(code === DialogDismissCode.ALL_CLOSED ? "All dialogs closed" : "Dialog closed");
        this.name = "DialogDismissedError";
        this.code = code;
    }
}

export const isDialogDismissedError = (error: unknown): boolean => {
    if (error instanceof DialogDismissedError) {
        return true;
    }

    if (!error || typeof error !== "object") {
        return false;
    }

    const code = (error as { code?: unknown }).code;
    return code === DialogDismissCode.CLOSED || code === DialogDismissCode.ALL_CLOSED;
};

const dialogs = ref<DialogInstance[]>([]);

export const useDialog = () => {
    const createDialog = (options: DialogOptions): Promise<any> => {
        return new Promise((resolve, reject) => {
            const dialog: DialogInstance = {
                id: Date.now() + Math.floor(Math.random() * 1000),
                visible: true,
                ...options,
                resolve,
                reject,
            };

            dialogs.value.push(dialog);
        });
    };

    const alert = (
        message: string,
        options?: Omit<DialogOptions, "message" | "type">,
    ) => {
        return createDialog({
            message,
            type: DialogType.ALERT,
            ...options,
        });
    };

    const confirm = (
        message: string,
        options?: Omit<DialogOptions, "message" | "type">,
    ) => {
        return createDialog({
            message,
            type: DialogType.CONFIRM,
            ...options,
        });
    };

    const prompt = (
        message: string,
        options?: Omit<DialogOptions, "message" | "type">,
    ) => {
        return createDialog({
            message,
            type: DialogType.PROMPT,
            ...options,
        });
    };

    const custom = (
        component: any,
        componentProps: Record<string, any> = {},
        options?: Omit<DialogOptions, "component" | "componentProps" | "type">,
    ) => {
        return createDialog({
            type: DialogType.CUSTOM,
            component: markRaw(component),
            componentProps,
            ...options,
        });
    };

    const closeDialog = (id: number, result?: any) => {
        const index = dialogs.value.findIndex((dialog) => dialog.id === id);
        if (index === -1) return;

        const dialog = dialogs.value[index];
        if (!dialog) return;

        if (result !== undefined && dialog.resolve) {
            dialog.resolve(result);
        } else if (dialog.reject) {
            dialog.reject(new DialogDismissedError(DialogDismissCode.CLOSED));
        }

        dialogs.value.splice(index, 1);
    };

    const closeAll = () => {
        dialogs.value.forEach((dialog) => {
            if (dialog.reject) {
                dialog.reject(new DialogDismissedError(DialogDismissCode.ALL_CLOSED));
            }
        });
        dialogs.value = [];
    };

    const handleDialogAction = (dialogId: number, action: DialogAction) => {
        const dialog = dialogs.value.find((d) => d.id === dialogId);
        if (!dialog) return;

        if (action.handler) {
            action.handler();
        }

        if (dialog.type === DialogType.CONFIRM) {
            if (action.primary) {
                closeDialog(dialogId, true);
            } else {
                // 非主要按钮视为取消
                closeDialog(dialogId, false);
            }
        } else if (dialog.type === DialogType.ALERT || dialog.type === DialogType.PROMPT) {
            if (action.primary) {
                closeDialog(dialogId, true);
            }
        }
    };

    const handleDialogClose = (dialogId: number) => {
        closeDialog(dialogId, false);
    };

    return {
        dialogs,
        alert,
        confirm,
        prompt,
        custom,
        closeDialog,
        closeAll,
        handleDialogAction,
        handleDialogClose,
    };
};