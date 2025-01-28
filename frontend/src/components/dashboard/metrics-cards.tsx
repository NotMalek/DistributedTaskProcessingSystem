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
            className: 'text-blue-500',
        },
        {
            title: 'Total Tasks',
            value: totalTasks,
            icon: Activity,
            className: 'text-gray-500',
        },
        {
            title: 'Processed Tasks',
            value: processedTasks,
            icon: CheckCircle,
            className: 'text-green-500',
        },
        {
            title: 'Failed Tasks',
            value: failedTasks,
            icon: XCircle,
            className: 'text-red-500',
        },
    ];

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {metrics.map((metric) => (
                <Card key={metric.title}>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">
                            {metric.title}
                        </CardTitle>
                        <metric.icon className={`h-4 w-4 ${metric.className}`} />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{metric.value}</div>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}