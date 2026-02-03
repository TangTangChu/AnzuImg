<template>
  <h1 class="mb-6 text-3xl font-bold text-center">
    {{ t("gallery.title") }}
  </h1>

  <div class="mx-auto mb-8 w-full max-w-4xl">
    <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,320px)] md:items-end">
      <div>
        <AnzuInput v-model="searchQuery" :placeholder="t('common.actions.search')" @keydown.enter="handleSearch">
          <template #prefix>
            <MagnifyingGlassIcon class="h-5 w-5" />
          </template>
        </AnzuInput>
      </div>

      <div>
        <div class="flex items-center gap-2">
          <AnzuComboBox v-model="selectedTag" :items="tagItems" :placeholder="t('tags.filterPlaceholder')"
            @change="handleTagChange" />
          <AnzuButton class="shrink-0 whitespace-nowrap" :disabled="!selectedTag" @click="clearTagFilter">
            {{ t("tags.clearFilter") }}
          </AnzuButton>
        </div>
      </div>
    </div>
  </div>

  <div v-if="pending" class="flex justify-center p-8">
    <AnzuProgressRing status="loading" />
  </div>

  <div v-else-if="error" class="p-4 text-center">
    <AnzuAlert type="error">{{ error.message }}</AnzuAlert>
  </div>
  <div v-else-if="!images?.data?.length" class="p-8 text-center text-(--md-sys-color-on-surface-variant)">
    {{ t("gallery.noImages") }}
  </div>
  <div v-else>
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
      <div v-for="(img, index) in images.data" :key="img.hash"
        class="group relative overflow-hidden rounded-xl bg-(--md-sys-color-surface-container) cursor-pointer"
        @click="openImageModal(img, index)">
        <img :src="`/i/${img.hash}/thumbnail`" :alt="img.file_name"
          class="aspect-square w-full object-cover transition-transform duration-300 group-hover:scale-105"
          loading="lazy" />
        <div
          class="absolute inset-0 flex flex-col justify-end bg-linear-to-t from-black/60 to-transparent p-3 opacity-0 transition-opacity duration-200 group-hover:opacity-100">
          <div class="flex items-center justify-end gap-2">
            <AnzuButton variant="filled" class="w-9! h-9! p-0! min-w-0! rounded-full shadow-sm"
              @click.stop="copyLink(img.hash)" :title="t('common.actions.copyLink')">
              <LinkIcon class="h-5 w-5" />
            </AnzuButton>
            <AnzuButton variant="tonal" class="w-9! h-9! p-0! min-w-0! rounded-full shadow-sm"
              @click.stop="deleteImage(img.hash)" :title="t('common.actions.delete')">
              <TrashIcon class="h-5 w-5" />
            </AnzuButton>
          </div>
        </div>
      </div>
    </div>
    <div class="mt-8 flex justify-center">
      <AnzuPagination :current-page="currentPage" :total-pages="totalPages" base-url="/gallery" />
    </div>
  </div>

  <ImageModal :image="currentImage" :visible="modalVisible" :current-index="currentImageIndex"
    :total-images="images?.data?.length || 0" :has-previous="hasPreviousImage" :has-next="hasNextImage"
    @update:visible="modalVisible = $event" @close="closeModal" @previous="showPreviousImage" @next="showNextImage"
    @copy-link="copyImageUrl" @download="downloadImage" @delete="deleteImage" />
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useAuth } from "~/composables/useAuth";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import AnzuAlert from "~/components/AnzuAlert.vue";
import ImageModal from "~/components/ImageModal.vue";
import AnzuPagination from "~/components/AnzuPagination.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import { useNotification } from "~/composables/useNotification";
import { useDialog } from "~/composables/useDialog";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import type { Image, ImageListResponse, TagListResponse } from "~/types/image";
import {
  LinkIcon,
  TrashIcon,
  MagnifyingGlassIcon,
} from "@heroicons/vue/24/outline";

const { t } = useI18n();
useAuth();
const { notify } = useNotification();
const { confirm } = useDialog();
const router = useRouter();
const route = useRoute();

const currentPage = computed(() => {
  const p = Number(route.query.page);
  return Number.isNaN(p) || p < 1 ? 1 : p;
});
const limit = 20;

const searchQuery = ref("");
const selectedTag = ref<string | null>(
  route.query.tag ? String(route.query.tag) : null,
);

if (route.query.file_name) {
  searchQuery.value = String(route.query.file_name);
} else if (route.query.tag) {
  searchQuery.value = "tag:" + String(route.query.tag);
}

const applyFilters = () => {
  const query: any = { page: 1 };
  const keyword = searchQuery.value.trim();

  if (keyword) {
    if (keyword.startsWith("tag:")) {
      const tagValue = keyword.substring(4).trim();
      if (tagValue) {
        query.tag = tagValue;
        selectedTag.value = tagValue;
      }
    } else {
      query.file_name = keyword;
    }
  }

  if (selectedTag.value) {
    query.tag = selectedTag.value;
  }

  router.push({ query });
};

const handleSearch = () => {
  applyFilters();
};

const handleTagChange = (value: string | number | null) => {
  const tagValue = value ? String(value) : null;
  selectedTag.value = tagValue;
  if (tagValue) {
    searchQuery.value = `tag:${tagValue}`;
  } else if (searchQuery.value.startsWith("tag:")) {
    searchQuery.value = "";
  }
  applyFilters();
};

const clearTagFilter = () => handleTagChange(null);

const {
  data: images,
  pending,
  error,
  refresh,
} = await useFetch<ImageListResponse>("/api/v1/images", {
  query: computed(() => ({
    page: currentPage.value,
    page_size: limit,
    file_name: route.query.file_name,
    tag: route.query.tag,
  })),
  watch: [currentPage, () => route.query],
});

const { data: tagList } = await useFetch<TagListResponse>("/api/v1/tags");
const tagItems = computed(() =>
  (tagList.value?.data ?? []).map((item) => ({
    value: item.tag,
    label: `${item.tag} (${item.count})`,
  })),
);

const totalPages = computed(() => {
  if (!images.value) return 0;
  const size = images.value.size || limit;
  return Math.ceil(images.value.total / size);
});

const modalVisible = ref(false);
const currentImage = ref<Image | null>(null);
const currentImageIndex = ref(0);
const imageLoaded = ref(false);

const hasPreviousImage = computed(() => {
  return !!(
    currentImageIndex.value > 0 &&
    images.value?.data &&
    images.value.data.length > 0
  );
});

const hasNextImage = computed(() => {
  return !!(
    images.value?.data && currentImageIndex.value < images.value.data.length - 1
  );
});

const openImageModal = (img: Image, index: number) => {
  currentImage.value = img;
  currentImageIndex.value = index;
  imageLoaded.value = false;
  modalVisible.value = true;
};

const closeModal = () => {
  modalVisible.value = false;
  currentImage.value = null;
};

const showPreviousImage = () => {
  if (hasPreviousImage.value && images.value?.data) {
    currentImageIndex.value--;
    const newImage = images.value.data[currentImageIndex.value];
    if (newImage) {
      currentImage.value = newImage;
      imageLoaded.value = false;
    }
  }
};

const showNextImage = () => {
  if (hasNextImage.value && images.value?.data) {
    currentImageIndex.value++;
    const newImage = images.value.data[currentImageIndex.value];
    if (newImage) {
      currentImage.value = newImage;
      imageLoaded.value = false;
    }
  }
};

const copyLink = (hash: string) => {
  const url = `${window.location.origin}/i/${hash}`;
  navigator.clipboard.writeText(url);
  notify({
    message: t("common.actions.copySuccess"),
    type: NotificationType.SUCCESS,
  });
};

const copyImageUrl = () => {
  if (currentImage.value) {
    copyLink(currentImage.value.hash);
  }
};

const downloadImage = () => {
  if (currentImage.value?.hash) {
    const url = `${window.location.origin}/i/${currentImage.value.hash}`;
    const a = document.createElement("a");
    a.href = url;
    a.download = currentImage.value.file_name || "image";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    notify({
      message: t("common.actions.deleteStarted"),
      type: NotificationType.SUCCESS,
    });
  }
};

const deleteImage = async (hash: string) => {
  try {
    const result = await confirm(t("common.actions.deleteConfirm"), {
      title: t("common.actions.delete"),
      variant: DialogVariant.DESTRUCTIVE,
      actions: [
        { text: t("common.actions.cancel"), variant: "text" },
        {
          text: t("common.actions.delete"),
          primary: true,
          variant: "filled",
          loading: false,
        },
      ],
    });

    if (!result) return;

    await $fetch(`/api/v1/images/${hash}`, {
      method: "DELETE",
    });
    notify({
      message: t("common.actions.deleteSuccess"),
      type: NotificationType.SUCCESS,
    });
    refresh();
    if (modalVisible.value && currentImage.value?.hash === hash) {
      closeModal();
    }
  } catch (e: any) {
    if (e.message === "Dialog closed" || e.message === "All dialogs closed") {
      return;
    }
    const errorMsg = e.data?.error || t("common.actions.deleteFailed");
    notify({
      message: errorMsg,
      type: NotificationType.ERROR,
    });
  }
};

const handleKeydown = (e: KeyboardEvent) => {
  if (!modalVisible.value) return;

  switch (e.key) {
    case "Escape":
      closeModal();
      break;
    case "ArrowLeft":
      if (hasPreviousImage.value) {
        e.preventDefault();
        showPreviousImage();
      }
      break;
    case "ArrowRight":
      if (hasNextImage.value) {
        e.preventDefault();
        showNextImage();
      }
      break;
  }
};

if (typeof window !== "undefined") {
  window.addEventListener("keydown", handleKeydown);
}
</script>
