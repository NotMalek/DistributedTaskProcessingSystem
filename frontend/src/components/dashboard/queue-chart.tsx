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

const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-[#27272a] border border-[#3f3f46] p-2 rounded-md shadow-lg">
                <p className="text-[#ec4899] text-sm font-medium">{`Tasks in Queue : ${payload[0].value}`}</p>
                <p className="text-gray-400 text-xs">{`Priority ${label}`}</p>
            </div>
        );
    }
    return null;
};

export function QueueChart({ queueLengths = {} }: QueueChartProps) {
    const data: QueueData[] = Object.entries(queueLengths || {}).map(([priority, length]) => ({
        priority: `P${priority}`,
        length: length || 0
    }));

    return (
        <Card className="bg-[#18181b] border-[#3f3f46]">
            <CardHeader>
                <CardTitle className="text-white">Queue Lengths by Priority</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                            <CartesianGrid
                                strokeDasharray="3 3"
                                stroke="#3f3f46"
                                vertical={false}
                            />
                            <XAxis
                                dataKey="priority"
                                stroke="#a1a1aa"
                                tick={{ fill: '#a1a1aa' }}
                            />
                            <YAxis
                                stroke="#a1a1aa"
                                tick={{ fill: '#a1a1aa' }}
                            />
                            <Tooltip
                                content={<CustomTooltip />}
                                cursor={{ fill: '#3f3f46' }}
                            />
                            <Legend
                                wrapperStyle={{ color: '#a1a1aa' }}
                            />
                            <Bar
                                dataKey="length"
                                fill="#ec4899"
                                name="Tasks in Queue"
                                radius={[4, 4, 0, 0]}
                            />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </CardContent>
        </Card>
    );
}