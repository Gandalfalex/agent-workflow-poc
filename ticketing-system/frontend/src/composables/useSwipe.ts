export type SwipeHandlers = {
    onSwipeLeft?: () => void;
    onSwipeRight?: () => void;
    onSwipeUp?: () => void;
    onSwipeDown?: () => void;
};

export function useSwipe(threshold = 60) {
    let startX = 0;
    let startY = 0;

    const onTouchStart = (e: TouchEvent) => {
        startX = e.touches[0]!.clientX;
        startY = e.touches[0]!.clientY;
    };

    const onTouchEnd = (e: TouchEvent, handlers: SwipeHandlers) => {
        const dx = e.changedTouches[0]!.clientX - startX;
        const dy = e.changedTouches[0]!.clientY - startY;
        if (Math.abs(dx) > Math.abs(dy)) {
            if (Math.abs(dx) >= threshold) {
                if (dx > 0) handlers.onSwipeRight?.();
                else handlers.onSwipeLeft?.();
            }
        } else {
            if (Math.abs(dy) >= threshold) {
                if (dy > 0) handlers.onSwipeDown?.();
                else handlers.onSwipeUp?.();
            }
        }
    };

    return { onTouchStart, onTouchEnd };
}
