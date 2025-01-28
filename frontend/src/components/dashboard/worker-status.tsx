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

    const inputClassName = "w-full p-2 rounded-md bg-[#27272a] text-white border border-[#3f3f46] focus:outline-none focus:border-[#ec4899] focus:ring-1 focus:ring-[#ec4899]";
    const labelClassName = "block text-sm font-medium mb-1 text-gray-200";

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

    return (
        <Card className="bg-[#18181b] border-[#3f3f46]">
            <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="text-white">Active Workers</CardTitle>
                <button
                    onClick={() => setShowNewWorkerForm(true)}
                    className="inline-flex items-center px-3 py-1.5 bg-[#ec4899] text-white rounded-md hover:bg-[#ec4899]/90"
                >
                    <Plus className="w-4 h-4 mr-2" />
                    Add Worker
                </button>
            </CardHeader>
            <CardContent>
                {error && (
                    <div className="mb-4 p-3 bg-red-900/20 text-red-400 rounded-md flex items-center border border-red-800">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        {error}
                    </div>
                )}

                {showNewWorkerForm && (
                    <Card className="mb-4 bg-[#1f1f23] border-[#3f3f46]">
                        <CardContent className="p-4">
                            <div className="flex justify-between items-center mb-4">
                                <h4 className="font-medium text-white">New Worker Configuration</h4>
                                <button
                                    onClick={() => setShowNewWorkerForm(false)}
                                    className="text-gray-400 hover:text-white"
                                >
                                    <X className="w-4 h-4" />
                                </button>
                            </div>
                            <div className="space-y-4">
                                <div>
                                    <label className={labelClassName}>
                                        Pool Size
                                    </label>
                                    <input
                                        type="number"
                                        value={poolSize}
                                        onChange={(e) => setPoolSize(Number(e.target.value))}
                                        className={inputClassName}
                                    />
                                </div>
                                <div>
                                    <label className="flex items-center space-x-2 text-gray-200">
                                        <input
                                            type="checkbox"
                                            checked={enableSteal}
                                            onChange={(e) => setEnableSteal(e.target.checked)}
                                            className="rounded bg-[#27272a] border-[#3f3f46] text-[#ec4899] focus:ring-[#ec4899]"
                                        />
                                        <span className="text-sm font-medium">Enable Task Stealing</span>
                                    </label>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className={labelClassName}>
                                            Min Workers
                                        </label>
                                        <input
                                            type="number"
                                            value={minWorkers}
                                            onChange={(e) => setMinWorkers(Number(e.target.value))}
                                            className={inputClassName}
                                        />
                                    </div>
                                    <div>
                                        <label className={labelClassName}>
                                            Max Workers
                                        </label>
                                        <input
                                            type="number"
                                            value={maxWorkers}
                                            onChange={(e) => setMaxWorkers(Number(e.target.value))}
                                            className={inputClassName}
                                        />
                                    </div>
                                </div>
                                <button
                                    onClick={startWorker}
                                    className="w-full py-2 px-4 bg-[#ec4899] text-white rounded-md hover:bg-[#ec4899]/90 transition-colors"
                                >
                                    Start Worker
                                </button>
                            </div>
                        </CardContent>
                    </Card>
                )}

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {Object.entries(workers).map(([id, worker]) => (
                        <Card key={id} className="bg-[#1f1f23] border-[#3f3f46]">
                            <CardContent className="p-4">
                                <div className="flex justify-between items-center mb-2">
                                    <span className="font-medium text-white">Worker {id.slice(0, 8)}</span>
                                    <span className={`px-2 py-1 rounded-full text-xs ${
                                        worker.status === 'active'
                                            ? 'bg-green-900/20 text-green-400 border border-green-800'
                                            : 'bg-yellow-900/20 text-yellow-400 border border-yellow-800'
                                    }`}>
                                        {worker.status}
                                    </span>
                                </div>
                                <div className="space-y-1 text-sm text-gray-300">
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
    );
}