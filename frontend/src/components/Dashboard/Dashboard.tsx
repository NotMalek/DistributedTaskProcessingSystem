import { useEffect, useState } from 'react';
import { SystemMetrics, fetchMetrics } from '../../lib/api';

function MetricCard({ title, value }: { title: string; value: number }) {
    return (
        <div className="bg-white rounded-lg shadow p-4">
            <h3 className="text-sm font-medium text-gray-500">{title}</h3>
            <p className="mt-2 text-3xl font-semibold text-gray-900">{value}</p>
        </div>
    );
}

export function Dashboard() {
    const [metrics, setMetrics] = useState<SystemMetrics>({
        active_workers: 0,
        total_tasks: 0,
        processed_tasks: 0,
        failed_tasks: 0,
        queue_lengths: {}
    });

    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const getMetrics = async () => {
            try {
                const data = await fetchMetrics();
                setMetrics(data);
                setError(null);
            } catch (err) {
                setError('Failed to fetch metrics');
            }
        };

        getMetrics();
        const interval = setInterval(getMetrics, 5000);
        return () => clearInterval(interval);
    }, []);

    if (error) {
        return <div className="text-red-500">Error: {error}</div>;
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <MetricCard title="Active Workers" value={metrics.active_workers} />
            <MetricCard title="Total Tasks" value={metrics.total_tasks} />
            <MetricCard title="Processed Tasks" value={metrics.processed_tasks} />
            <MetricCard title="Failed Tasks" value={metrics.failed_tasks} />
        </div>
    );
}