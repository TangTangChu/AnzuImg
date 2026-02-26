import { useApi } from "~/composables/useApi";
import type { SystemStats } from "~/types/stats";

export const useStats = () => {
    const { apiUrl } = useApi();

    const fetchStats = async () => {
        try {
            const data = await $fetch<SystemStats>(apiUrl('/api/v1/stats'));
            return data;
        } catch (error) {
            console.error('Fetch stats failed', error);
            return null;
        }
    };

    return {
        fetchStats,
    };
};
