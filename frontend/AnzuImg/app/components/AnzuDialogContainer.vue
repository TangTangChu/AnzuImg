<template>
    <div>
        <AnzuDialog v-for="dialog in dialogs" :key="dialog.id" :title="dialog.title" :message="dialog.message"
            :type="dialog.type" :size="dialog.size" :position="dialog.position" :variant="dialog.variant"
            :actions="dialog.actions" :show-close-button="dialog.showCloseButton"
            :close-on-click-outside="dialog.closeOnClickOutside" :close-on-esc="dialog.closeOnEsc"
            :show-overlay="dialog.showOverlay" :overlay-clickable="dialog.overlayClickable"
            :persistent="dialog.persistent" :max-width="dialog.maxWidth" :min-width="dialog.minWidth"
            :max-height="dialog.maxHeight" :min-height="dialog.minHeight" :custom-class="dialog.customClass"
            :icon="dialog.icon" :component="dialog.component" :component-props="dialog.componentProps"
            :visible="dialog.visible" @update:visible="handleVisibilityChange(dialog.id, $event)"
            @close="handleDialogClose(dialog.id)" @action="handleDialogAction(dialog.id, $event)"
            @confirm="handleDialogConfirm(dialog.id)" @cancel="handleDialogCancel(dialog.id)" />
    </div>
</template>

<script setup lang="ts">
import { useDialog } from "@/composables/useDialog";
import AnzuDialog from "./AnzuDialog.vue";

const {
    dialogs,
    handleDialogAction: handleAction,
    handleDialogClose: handleClose,
} = useDialog();

const handleVisibilityChange = (id: number, visible: boolean) => {
    if (!visible) {
        handleClose(id);
    }
};

const handleDialogAction = (id: number, action: any) => {
    handleAction(id, action);
};

const handleDialogClose = (id: number) => {
    handleClose(id);
};

const handleDialogConfirm = (id: number) => {
    handleClose(id);
};

const handleDialogCancel = (id: number) => {
    handleClose(id);
};
</script>