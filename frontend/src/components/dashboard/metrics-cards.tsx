import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Activity, Users, CheckCircle, XCircle } from 'lucide-react';

interface MetricsCardsProps {
    activeWorkers: number;
    totalTasks: number;
    processedTasks: number;
    failedTasks: number;
}

export function MetricsCards({
                                 activeWorkers,
                                 totalTasks,
                                 processedTasks,
                                 failedTasks,
                             }: MetricsCardsProps) {
    const metrics = [
        {
            title: 'Active Workers',
            value: activeWorkers,
            icon: Users,
            color: 'task-highlight',
            description: 'Currently running worker instances'
        },
        {
            title: 'Total Tasks',
            value: totalTasks,
            icon: Activity,
            color: 'muted-foreground',
            description: 'Tasks in all queues'
        },
        {
            title: 'Processed Tasks',
            value: processedTasks,
            icon: CheckCircle,
            color: 'success',
            description: 'Successfully completed tasks'
        },
        {
            title: 'Failed Tasks',
            value: failedTasks,
            icon: XCircle,
            color: 'error',
            description: 'Tasks that encountered errors'
        },
    ];

    return (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
            {metrics.map((metric) => (
                <Card
                    key={metric.title}
                    className="relative overflow-hidden"
                    style={{
                        backgroundColor: 'var(--card-background)',
                        color: 'var(--card-foreground)'
                    }}
                >
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">
                            {metric.title}
                        </CardTitle>
                        <metric.icon
                            className="h-4 w-4"
                            style={{ color: `var(--${metric.color})` }}
                        />
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-col space-y-2">
                            <span className="text-2xl font-bold">
                                {metric.value.toLocaleString()}
                            </span>
                            <p
                                className="text-xs"
                                style={{ color: 'var(--muted-foreground)' }}
                            >
                                {metric.description}
                            </p>
                        </div>
                        <div
                            className="absolute bottom-0 left-0 h-1 w-full"
                            style={{
                                backgroundColor: `var(--${metric.color})`,
                                opacity: '0.2'
                            }}
                        />
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}