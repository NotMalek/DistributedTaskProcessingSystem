export interface SystemMetrics {
    active_workers: number;
    total_tasks: number;
    processed_tasks: number;
    failed_tasks: number;
    queue_lengths: Record<number, number>;
}

export async function fetchMetrics(): Promise<SystemMetrics> {
    // Mock data for testing
    return {
        active_workers: 5,
        total_tasks: 100,
        processed_tasks: 75,
        failed_tasks: 2,
        queue_lengths: {
            1: 10,
            2: 8,
            3: 5,
            4: 2
        }
    };
}