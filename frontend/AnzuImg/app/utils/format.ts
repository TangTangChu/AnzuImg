export const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KiB", "MiB", "GiB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

export const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second:"2-digit"
    });
};

export const formatRelativeTime = (dateString: string, locale: string = 'zh-CN'): string => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    try {
        const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
        
        if (diffInSeconds < 60) return rtf.format(-diffInSeconds, 'second');
        if (diffInSeconds < 3600) return rtf.format(-Math.floor(diffInSeconds / 60), 'minute');
        if (diffInSeconds < 86400) return rtf.format(-Math.floor(diffInSeconds / 3600), 'hour');
        if (diffInSeconds < 604800) return rtf.format(-Math.floor(diffInSeconds / 86400), 'day'); // 7 days
        if (diffInSeconds < 2592000) return rtf.format(-Math.floor(diffInSeconds / 604800), 'week'); // 30 days roughly
        if (diffInSeconds < 31536000) return rtf.format(-Math.floor(diffInSeconds / 2592000), 'month');
        return rtf.format(-Math.floor(diffInSeconds / 31536000), 'year');
    } catch (e) {
        return date.toLocaleDateString();
    }
};
