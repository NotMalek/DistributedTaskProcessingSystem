import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface WorkerMetrics {
    id: string;
    tasks_processed: number;
    cpu_usage: number;
    memory_usage: number;
    status: string;
}

interface WorkerStatusProps {
    workers?: Record<string, WorkerMetrics>;
}

export function WorkerStatus({ workers = {} }: WorkerStatusProps) {
    const workerData = Object.values(workers || {}).map(worker => ({
        id: worker?.id?.slice(0, 8) || 'unknown',
        tasksProcessed: worker?.tasks_processed || 0,
        cpuUsage: worker?.cpu_usage || 0,
        memoryUsage: (worker?.memory_usage || 0) / (1024 * 1024), // Convert to MB
        status: worker?.status || 'unknown'
    }));

    return (
        <div className="space-y-6">
            <Card className="p-4">
                <CardHeader>
                    <CardTitle>Worker Performance</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="h-64">
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={workerData}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="id" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Line type="monotone" dataKey="cpuUsage" stroke="#3b82f6" name="CPU Usage %" />
                                <Line type="monotone" dataKey="memoryUsage" stroke="#10b981" name="Memory Usage (MB)" />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Active Workers</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {workerData.map((worker) => (
                            <Card key={worker.id} className="bg-muted">
                                <CardContent className="p-4">
                                    <div className="flex justify-between items-center mb-2">
                                        <span className="font-medium">Worker {worker.id}</span>
                                        <span className={`px-2 py-1 rounded-full text-xs ${
                                            worker.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                                        }`}>
                      {worker.status}
                    </span>
                                    </div>
                                    <div className="space-y-1 text-sm">
                                        <div className="flex justify-between">
                                            <span>Tasks Processed:</span>
                                            <span>{worker.tasksProcessed}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>CPU Usage:</span>
                                            <span>{worker.cpuUsage.toFixed(1)}%</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Memory:</span>
                                            <span>{worker.memoryUsage.toFixed(1)} MB</span>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}