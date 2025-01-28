import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface QueueData {
    priority: string;
    length: number;
}

interface QueueChartProps {
    queueLengths?: Record<string, number>;
}

export function QueueChart({ queueLengths = {} }: QueueChartProps) {
    const data: QueueData[] = Object.entries(queueLengths || {}).map(([priority, length]) => ({
        priority: `P${priority}`,
        length: length || 0
    }));

    return (
        <Card className="p-4">
            <CardHeader>
                <CardTitle>Queue Lengths by Priority</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={data}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="priority" />
                            <YAxis />
                            <Tooltip />
                            <Legend />
                            <Bar dataKey="length" fill="var(--task-highlight)" name="Tasks in Queue" />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </CardContent>
        </Card>
    );
}