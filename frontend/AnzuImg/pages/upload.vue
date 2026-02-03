<template>
    <div class="h-full flex flex-col max-w-6xl mx-auto w-full">
        <h1 class="mb-6 text-3xl font-bold text-center">{{ t("upload.title") }}</h1>
        <div v-if="files.length === 0"
            class="flex-1 flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-(--md-sys-color-outline-variant) transition-colors min-h-100 relative group p-12 cursor-pointer"
            :class="[isDragging ? 'border-(--md-sys-color-primary)' : '']" @dragenter.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false" @dragover.prevent @drop.prevent="handleDrop"
            @click="triggerMainInput">
            <input type="file" ref="fileInput" class="hidden" @change="handleFileSelect" accept="image/*" multiple />

            <svg class="mx-auto mb-4 h-16 w-16 text-(--md-sys-color-primary)" fill="none" stroke="currentColor"
                viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <p class="mb-2 text-xl font-medium text-(--md-sys-color-on-surface)">
                {{ t("upload.dragDrop") }}
            </p>
            <p class="text-sm text-(--md-sys-color-on-surface-variant)">
                {{ t("upload.orSelect") }}
            </p>
        </div>
        <div v-else class="flex flex-col lg:flex-row gap-6 flex-1 min-h-0 animate-fade-in-up">
            <div
                class="lg:w-1/2 flex flex-col min-h-125 lg:h-150 rounded-xl overflow-hidden border border-(--md-sys-color-outline-variant)">
                <div class="p-4 border-b border-(--md-sys-color-outline-variant) flex justify-between items-center">
                    <span class="font-medium">{{ files.length }} {{ t("common.labels.files") }} ({{
                        formatFileSize(totalSize)
                    }})</span>
                    <AnzuButton variant="text" @click="clearAll">{{
                        t("common.actions.clear")
                    }}</AnzuButton>
                </div>

                <div class="flex-1 overflow-y-auto p-4">
                    <div class="grid grid-cols-3 sm:grid-cols-4 gap-3">
                        <div v-for="(item, index) in files" :key="index"
                            class="relative aspect-square rounded-lg overflow-hidden border-2 cursor-pointer transition-all group"
                            :class="[
                                selectedIndex === index
                                    ? 'border-(--md-sys-color-primary) ring-2 ring-(--md-sys-color-primary/20)'
                                    : 'border-transparent hover:border-(--md-sys-color-outline)',
                                item.status === 'error' ? 'border-red-500!' : '',
                                item.status === 'success' ? 'border-green-500!' : '',
                            ]" @click="selectFile(index)">
                            <img :src="item.previewUrl" class="w-full h-full object-cover" />

                            <div class="absolute inset-0 flex items-center justify-center bg-black/40"
                                v-if="item.status !== 'pending'">
                                <div v-if="item.status === 'success'" class="text-green-400">
                                    <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M5 13l4 4L19 7" />
                                    </svg>
                                </div>
                                <div v-if="item.status === 'error'" class="text-red-400">
                                    <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </div>
                            </div>
                            <div class="absolute inset-0 bg-primary/10 opacity-0 group-hover:opacity-100 transition-opacity"
                                :class="{ 'opacity-100': selectedIndex === index }"></div>
                        </div>
                        <div class="relative aspect-square rounded-lg border-2 border-dashed border-(--md-sys-color-outline-variant) flex items-center justify-center cursor-pointer hover:bg-(--md-sys-color-surface-container) hover:border-(--md-sys-color-primary) transition-colors"
                            @click="triggerAddInput">
                            <input type="file" ref="addInput" class="hidden" @change="handleAddFile" accept="image/*"
                                multiple />
                            <svg class="w-8 h-8 text-(--md-sys-color-on-surface-variant)" fill="none"
                                stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M12 4v16m8-8H4" />
                            </svg>
                        </div>
                    </div>
                </div>
            </div>
            <div
                class="lg:w-1/2 flex flex-col h-auto lg:h-150 rounded-xl border border-(--md-sys-color-outline-variant) overflow-hidden">
                <div v-if="selectedFile" class="flex-1 flex flex-col min-h-0">
                    <div class="p-4 border-b border-(--md-sys-color-outline-variant) flex gap-4 items-center">
                        <div
                            class="h-16 w-16 rounded overflow-hidden shrink-0 border border-(--md-sys-color-outline-variant)">
                            <img :src="selectedFile.previewUrl" class="w-full h-full object-contain" />
                        </div>
                        <div class="flex-1 min-w-0">
                            <h3 class="font-bold truncate text-(--md-sys-color-on-surface)">
                                {{ selectedFile.file.name }}
                            </h3>
                            <p class="text-xs text-(--md-sys-color-on-surface-variant)">
                                {{ formatFileSize(selectedFile.file.size) }} ·
                                {{ selectedFile.file.type }}
                            </p>
                        </div>
                    </div>
                    <div class="flex-1 overflow-y-auto p-6 space-y-4">
                        <AnzuInput v-model="selectedFile.customName" :label="t('upload.customFileName')"
                            :placeholder="t('upload.customFileNamePlaceholder')" />

                        <AnzuInput v-model="selectedFile.description" :label="t('common.labels.description')" />

                        <div class="flex items-center gap-2">
                            <AnzuComboBox v-model="selectedTagOption" :items="tagItems"
                                :placeholder="t('tags.selectPlaceholder')" :aria-label="t('tags.selectLabel')"
                                @change="handleTagPick" />
                            <AnzuButton class="shrink-0 whitespace-nowrap" variant="tonal"
                                :disabled="!selectedTagOption" @click="addSelectedTag">
                                {{ t("tags.add") }}
                            </AnzuButton>
                        </div>

                        <AnzuTags v-model="selectedFile.tags" :label="t('common.labels.tags')" :max-tags="10" />

                        <AnzuTags v-model="selectedFile.routes" :label="t('upload.route')" :max-tags="5" />
                    </div>

                    <div v-if="selectedFile.status === 'success'"
                        class="p-4 border-t border-(--md-sys-color-outline-variant) bg-(--md-sys-color-primary-container)/40 text-(--md-sys-color-on-primary-container)">
                        <p class="font-bold text-sm flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M5 13l4 4L19 7" />
                            </svg>
                            {{ t("upload.success") }}
                        </p>
                        <a :href="selectedFile.resultUrl" target="_blank"
                            class="text-xs underline break-all mt-1 block">{{ selectedFile.resultUrl }}</a>
                    </div>
                    <div v-if="selectedFile.status === 'error'"
                        class="p-4 bg-red-500/10 border-t border-red-500/20 text-red-600">
                        <p class="font-bold text-sm flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                            </svg>
                            Upload Failed
                        </p>
                        <p class="text-xs mt-1">{{ selectedFile.error }}</p>
                    </div>
                </div>
                <div v-else class="flex-1 flex items-center justify-center text-(--md-sys-color-on-surface-variant)">
                    Select an image to edit details
                </div>
                <div class="p-4 border-t border-(--md-sys-color-outline-variant)">
                    <div class="mb-4 p-3 rounded-lg border border-(--md-sys-color-outline-variant)">
                        <div class="flex items-center gap-2 mb-2">
                            <AnzuCheckbox v-model="enableConvert" :label="t('upload.convert') + ' (All)'" />
                        </div>
                        <div v-if="enableConvert" class="grid grid-cols-3 gap-2 text-sm animate-fade-in-up">
                            <div>
                                <label
                                    class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1">Format</label>
                                <AnzuComboBox v-model="targetFormat" :items="['webp', 'avif']" />
                            </div>
                            <div>
                                <label
                                    class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1">{{ t("upload.quality") }}</label>
                                <AnzuInput v-model="quality" type="number" placeholder="80" />
                            </div>
                            <div>
                                <label
                                    class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1">{{ t("upload.effort") }}</label>
                                <AnzuInput v-model="effort" type="number" placeholder="4" />
                            </div>
                        </div>
                    </div>
                    <AnzuButton @click="startUpload" :status="uploading ? 'loading' : 'default'" class="w-full"
                        :disabled="uploading || files.length === 0">
                        {{ t("upload.submit") }} ({{ files.length }})
                    </AnzuButton>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from "vue";
import { useAuth } from "~/composables/useAuth";
import { formatFileSize } from "~/utils/format";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuCheckbox from "~/components/AnzuCheckbox.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";
import type { TagListResponse } from "~/types/image";

const { t } = useI18n();
useAuth();
const { notify } = useNotification();

interface UploadFileItem {
    file: File;
    previewUrl: string;
    description: string;
    tags: string[];
    routes: string[];
    customName: string;
    status: "pending" | "success" | "error";
    error?: string;
    resultUrl?: string;
}

const isDragging = ref(false);
const uploading = ref(false);
const files = ref<UploadFileItem[]>([]);
const selectedIndex = ref<number>(0);
const fileInput = ref<HTMLInputElement | null>(null);
const addInput = ref<HTMLInputElement | null>(null);

const enableConvert = ref(false);
const targetFormat = ref("webp");
const quality = ref("");
const effort = ref("");

const selectedTagOption = ref<string | null>(null);
const { data: tagList } = await useFetch<TagListResponse>("/api/v1/tags");
const tagItems = computed(() =>
    (tagList.value?.data ?? []).map((item) => ({
        value: item.tag,
        label: `${item.tag} (${item.count})`,
    })),
);

const selectedFile = computed(() => {
    if (files.value.length === 0) return null;
    return files.value[selectedIndex.value];
});

const handleTagPick = (value: string | number | null) => {
    selectedTagOption.value = value ? String(value) : null;
};

const addSelectedTag = () => {
    if (!selectedFile.value || !selectedTagOption.value) return;
    if (!selectedFile.value.tags.includes(selectedTagOption.value)) {
        selectedFile.value.tags.push(selectedTagOption.value);
    }
    selectedTagOption.value = null;
};

const totalSize = computed(() => {
    return files.value.reduce((acc, item) => acc + item.file.size, 0);
});

onUnmounted(() => {
    files.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
});

const triggerMainInput = () => {
    fileInput.value?.click();
};

const triggerAddInput = () => {
    addInput.value?.click();
};

const processFiles = (newFiles: FileList | null) => {
    if (!newFiles) return;
    const newItems = Array.from(newFiles).map((file) => ({
        file,
        previewUrl: URL.createObjectURL(file),
        description: "",
        tags: [],
        routes: [],
        customName: "",
        status: "pending" as const,
    }));

    const startIndex = files.value.length;
    files.value = [...files.value, ...newItems];
    if (newItems.length > 0) {
        selectedIndex.value = startIndex;
    }
};

const handleDrop = (e: DragEvent) => {
    isDragging.value = false;
    processFiles(e.dataTransfer?.files || null);
};

const handleFileSelect = (e: Event) => {
    const input = e.target as HTMLInputElement;
    processFiles(input.files);
    input.value = "";
};

const handleAddFile = (e: Event) => {
    const input = e.target as HTMLInputElement;
    processFiles(input.files);
    input.value = "";
};

const selectFile = (index: number) => {
    selectedIndex.value = index;
};

const clearAll = () => {
    files.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
    files.value = [];
    selectedIndex.value = 0;
};

const startUpload = async () => {
    if (files.value.length === 0) return;

    uploading.value = true;

    // Reset
    files.value.forEach((f) => {
        if (f.status === "error") f.status = "pending";
    });

    const formData = new FormData();
    // Append all
    files.value.forEach((item) => {
        formData.append("file", item.file);
    });

    // Append metadata
    const metadata = files.value.map((f) => ({
        description: f.description,
        tags: f.tags,
        routes: f.routes,
        custom_name: f.customName,
    }));
    formData.append("metadata", JSON.stringify(metadata));

    // Global settings
    if (enableConvert.value) {
        formData.append("convert", "true");
        formData.append("target_format", targetFormat.value);
        if (quality.value) formData.append("quality", quality.value);
        if (effort.value) formData.append("effort", effort.value);
    }

    try {
        const data = await $fetch<any[]>("/api/v1/images", {
            method: "POST",
            body: formData,
        });

        if (Array.isArray(data)) {
            let successCount = 0;
            data.forEach((res, index) => {
                if (files.value[index]) {
                    if (res.success) {
                        files.value[index].status = "success";
                        files.value[index].resultUrl =
                            `${window.location.origin}${res.url}`;
                        successCount++;
                    } else {
                        files.value[index].status = "error";
                        files.value[index].error = res.error || "Unknown error";
                    }
                }
            });

            if (successCount === files.value.length) {
                notify({
                    message: t("upload.success"),
                    type: NotificationType.SUCCESS,
                });
            } else {
                notify({
                    message: `Uploaded with ${files.value.length - successCount} errors`,
                    type: NotificationType.WARNING,
                });
            }
        }
    } catch (e: any) {
        const errorMsg = e.data && e.data.error ? e.data.error : "Upload failed";
        notify({
            message: errorMsg,
            type: NotificationType.ERROR,
        });
    } finally {
        uploading.value = false;
    }
};
</script>

<style scoped>
@reference "tailwindcss";

.animate-fade-in-up {
    animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
    from {
        opacity: 0;
        transform: translateY(10px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>
