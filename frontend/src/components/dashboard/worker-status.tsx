import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { AlertCircle, Plus, X } from 'lucide-react';

interface WorkerMetrics {
    id: string;
    lastSeen: string;
    tasksProcessed: number;
    activeTasks: number;
    status: string;
}

interface WorkerStatusProps {
    workers?: Record<string, WorkerMetrics>;
}

export function WorkerStatus({ workers = {} }: WorkerStatusProps) {
    const [showNewWorkerForm, setShowNewWorkerForm] = useState(false);
    const [poolSize, setPoolSize] = useState(5);
    const [enableSteal, setEnableSteal] = useState(true);
    const [minWorkers, setMinWorkers] = useState(1);
    const [maxWorkers, setMaxWorkers] = useState(10);
    const [error, setError] = useState<string | null>(null);

    const startWorker = async () => {
        try {
            const response = await fetch('http://localhost:8080/api/workers/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    poolSize,
                    enableSteal,
                    minWorkers,
                    maxWorkers,
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to start worker');
            }

            setShowNewWorkerForm(false);
            setError(null);
        } catch (err) {
            setError('Failed to start worker');
        }
    };

    const stopWorker = async (workerId: string) => {
        try {
            const response = await fetch(`http://localhost:8080/api/workers/stop?id=${workerId}`, {
                method: 'POST',
            });

            if (!response.ok) {
                throw new Error('Failed to stop worker');
            }

            setError(null);
        } catch (err) {
            setError('Failed to stop worker');
        }
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle>Active Workers</CardTitle>
                    <button
                        onClick={() => setShowNewWorkerForm(true)}
                        className="inline-flex items-center px-3 py-1 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                    >
                        <Plus className="w-4 h-4 mr-2" />
                        Add Worker
                    </button>
                </CardHeader>
                <CardContent>
                    {error && (
                        <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-md flex items-center">
                            <AlertCircle className="w-4 h-4 mr-2" />
                            {error}
                        </div>
                    )}

                    {showNewWorkerForm && (
                        <Card className="mb-4 bg-gray-50">
                            <CardContent className="p-4">
                                <div className="flex justify-between items-center mb-4">
                                    <h4 className="font-medium">New Worker Configuration</h4>
                                    <button
                                        onClick={() => setShowNewWorkerForm(false)}
                                        className="text-gray-500 hover:text-gray-700"
                                    >
                                        <X className="w-4 h-4" />
                                    </button>
                                </div>
                                <div className="space-y-4">
                                    <div>
                                        <label className="block text-sm font-medium mb-1">
                                            Pool Size
                                        </label>
                                        <input
                                            type="number"
                                            value={poolSize}
                                            onChange={(e) => setPoolSize(Number(e.target.value))}
                                            className="w-full p-2 border rounded-md"
                                        />
                                    </div>
                                    <div>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={enableSteal}
                                                onChange={(e) => setEnableSteal(e.target.checked)}
                                                className="rounded"
                                            />
                                            <span className="text-sm font-medium">Enable Task Stealing</span>
                                        </label>
                                    </div>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm font-medium mb-1">
                                                Min Workers
                                            </label>
                                            <input
                                                type="number"
                                                value={minWorkers}
                                                onChange={(e) => setMinWorkers(Number(e.target.value))}
                                                className="w-full p-2 border rounded-md"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium mb-1">
                                                Max Workers
                                            </label>
                                            <input
                                                type="number"
                                                value={maxWorkers}
                                                onChange={(e) => setMaxWorkers(Number(e.target.value))}
                                                className="w-full p-2 border rounded-md"
                                            />
                                        </div>
                                    </div>
                                    <button
                                        onClick={startWorker}
                                        className="w-full py-2 px-4 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                                    >
                                        Start Worker
                                    </button>
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {Object.entries(workers).map(([id, worker]) => (
                            <Card key={id} className="bg-gray-50">
                                <CardContent className="p-4">
                                    <div className="flex justify-between items-center mb-2">
                                        <span className="font-medium">Worker {id.slice(0, 8)}</span>
                                        <div className="flex items-center space-x-2">
                                            <span className={`px-2 py-1 rounded-full text-xs ${
                                                worker.status === 'active'
                                                    ? 'bg-green-100 text-green-800'
                                                    : 'bg-yellow-100 text-yellow-800'
                                            }`}>
                                                {worker.status}
                                            </span>
                                            <button
                                                onClick={() => stopWorker(id)}
                                                className="p-1 hover:bg-gray-200 rounded"
                                            >
                                                <X className="w-4 h-4 text-red-500" />
                                            </button>
                                        </div>
                                    </div>
                                    <div className="space-y-1 text-sm">
                                        <div className="flex justify-between">
                                            <span>Tasks Processed:</span>
                                            <span>{worker.tasksProcessed}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Active Tasks:</span>
                                            <span>{worker.activeTasks}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Last Seen:</span>
                                            <span>{new Date(worker.lastSeen).toLocaleTimeString()}</span>
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