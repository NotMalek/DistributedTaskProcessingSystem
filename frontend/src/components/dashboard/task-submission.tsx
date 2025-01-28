import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { AlertCircle } from 'lucide-react';

interface TaskSubmissionFormProps {
    onSuccess?: () => void;
}

export function TaskSubmissionForm({ onSuccess }: TaskSubmissionFormProps) {
    const [priority, setPriority] = useState(5);
    const [deadline, setDeadline] = useState('');
    const [retries, setRetries] = useState(3);
    const [taskType, setTaskType] = useState('test');
    const [payload, setPayload] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const [taskId, setTaskId] = useState<string | null>(null);
    const [taskStatus, setTaskStatus] = useState<any | null>(null);

    const submitTask = async () => {
        setSubmitting(true);
        setError(null);

        try {
            const response = await fetch('http://localhost:8080/api/tasks/submit', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    priority,
                    deadline: deadline || undefined,
                    retries,
                    taskType,
                    payload,
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to submit task');
            }

            const result = await response.json();
            setTaskId(result.taskId);
            onSuccess?.();

            // Start polling for task status
            pollTaskStatus(result.taskId);
        } catch (err) {
            setError('Failed to submit task');
        } finally {
            setSubmitting(false);
        }
    };

    const pollTaskStatus = async (id: string) => {
        const checkStatus = async () => {
            try {
                const response = await fetch(`http://localhost:8080/api/tasks/status?id=${id}`);
                if (response.ok) {
                    const status = await response.json();
                    setTaskStatus(status);
                    if (status.status === 'completed' || status.status === 'failed') {
                        return true;
                    }
                }
                return false;
            } catch (err) {
                return false;
            }
        };

        const poll = async () => {
            const isComplete = await checkStatus();
            if (!isComplete) {
                setTimeout(poll, 1000);
            }
        };

        poll();
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>Submit New Task</CardTitle>
            </CardHeader>
            <CardContent>
                {error && (
                    <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-md flex items-center">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        {error}
                    </div>
                )}

                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium mb-1">
                            Priority (1-10)
                        </label>
                        <input
                            type="number"
                            min="1"
                            max="10"
                            value={priority}
                            onChange={(e) => setPriority(Number(e.target.value))}
                            className="w-full p-2 border rounded-md"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium mb-1">
                            Deadline (optional)
                        </label>
                        <input
                            type="datetime-local"
                            value={deadline}
                            onChange={(e) => setDeadline(e.target.value)}
                            className="w-full p-2 border rounded-md"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium mb-1">
                            Max Retries
                        </label>
                        <input
                            type="number"
                            min="0"
                            value={retries}
                            onChange={(e) => setRetries(Number(e.target.value))}
                            className="w-full p-2 border rounded-md"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium mb-1">
                            Task Type
                        </label>
                        <input
                            type="text"
                            value={taskType}
                            onChange={(e) => setTaskType(e.target.value)}
                            className="w-full p-2 border rounded-md"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium mb-1">
                            Payload
                        </label>
                        <textarea
                            value={payload}
                            onChange={(e) => setPayload(e.target.value)}
                            className="w-full p-2 border rounded-md h-24"
                            placeholder="Enter task data..."
                        />
                    </div>

                    <button
                        onClick={submitTask}
                        disabled={submitting}
                        className="w-full py-2 px-4 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-blue-300"
                    >
                        {submitting ? 'Submitting...' : 'Submit Task'}
                    </button>

                    {taskId && (
                        <div className="mt-4 p-4 bg-gray-50 rounded-md">
                            <h4 className="font-medium mb-2">Task Status</h4>
                            <div className="text-sm">
                                <div className="flex justify-between mb-1">
                                    <span>Task ID:</span>
                                    <span className="font-mono">{taskId}</span>
                                </div>
                                {taskStatus && (
                                    <>
                                        <div className="flex justify-between mb-1">
                                            <span>Status:</span>
                                            <span className={
                                                taskStatus.status === 'completed'
                                                    ? 'text-green-600'
                                                    : taskStatus.status === 'failed'
                                                        ? 'text-red-600'
                                                        : 'text-blue-600'
                                            }>
                                                {taskStatus.status}
                                            </span>
                                        </div>
                                        {taskStatus.result && (
                                            <div className="mt-2 p-2 bg-gray-100 rounded">
                                                <div className="font-medium mb-1">Result:</div>
                                                <pre className="text-xs overflow-auto">
                                                    {JSON.stringify(taskStatus.result, null, 2)}
                                                </pre>
                                            </div>
                                        )}
                                        {taskStatus.error && (
                                            <div className="mt-2 p-2 bg-red-50 text-red-700 rounded">
                                                <div className="font-medium mb-1">Error:</div>
                                                <pre className="text-xs overflow-auto">
                                                    {taskStatus.error}
                                                </pre>
                                            </div>
                                        )}
                                    </>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}