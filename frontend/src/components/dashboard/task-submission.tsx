import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { AlertCircle, Pen, Camera } from 'lucide-react';

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

            onSuccess?.();
            // Reset form
            setPriority(5);
            setDeadline('');
            setRetries(3);
            setTaskType('test');
            setPayload('');
        } catch (err) {
            setError('Failed to submit task');
        } finally {
            setSubmitting(false);
        }
    };

    const inputClassName = "w-full p-2 rounded-md bg-[#27272a] text-white border border-[#3f3f46] focus:outline-none focus:border-[#ec4899] focus:ring-1 focus:ring-[#ec4899]";
    const labelClassName = "block text-sm font-medium mb-1 text-gray-200";

    return (
        <Card className="bg-[#18181b] border-[#3f3f46]">
            <CardHeader>
                <CardTitle className="text-white">Submit New Task</CardTitle>
            </CardHeader>
            <CardContent>
                {error && (
                    <div className="mb-4 p-3 bg-red-900/20 text-red-400 rounded-md flex items-center border border-red-800">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        {error}
                    </div>
                )}

                <div className="space-y-4">
                    <div>
                        <label className={labelClassName}>
                            Priority (1-10)
                        </label>
                        <input
                            type="number"
                            min="1"
                            max="10"
                            value={priority}
                            onChange={(e) => setPriority(Number(e.target.value))}
                            className={inputClassName}
                        />
                    </div>

                    <div>
                        <label className={labelClassName}>
                            Deadline (optional)
                        </label>
                        <input
                            type="datetime-local"
                            value={deadline}
                            onChange={(e) => setDeadline(e.target.value)}
                            className={inputClassName}
                        />
                    </div>

                    <div>
                        <label className={labelClassName}>
                            Max Retries
                        </label>
                        <input
                            type="number"
                            min="0"
                            value={retries}
                            onChange={(e) => setRetries(Number(e.target.value))}
                            className={inputClassName}
                        />
                    </div>

                    <div>
                        <label className={labelClassName}>
                            Task Type
                        </label>
                        <input
                            type="text"
                            value={taskType}
                            onChange={(e) => setTaskType(e.target.value)}
                            className={inputClassName}
                        />
                    </div>

                    <div>
                        <label className={labelClassName}>
                            Payload
                        </label>
                        <div className="relative">
                            <textarea
                                value={payload}
                                onChange={(e) => setPayload(e.target.value)}
                                className={`${inputClassName} h-24 resize-none`}
                                placeholder="Enter task data..."
                            />
                            <div className="absolute bottom-2 right-2 flex space-x-2">
                                <button className="p-1.5 rounded bg-[#3f3f46] hover:bg-[#52525b] text-white">
                                    <Pen className="w-4 h-4" />
                                </button>
                                <button className="p-1.5 rounded bg-[#3f3f46] hover:bg-[#52525b] text-white">
                                    <Camera className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    </div>

                    <button
                        onClick={submitTask}
                        disabled={submitting}
                        className="w-full py-2 px-4 bg-[#ec4899] text-white rounded-md hover:bg-[#ec4899]/90 disabled:bg-[#ec4899]/50 transition-colors"
                    >
                        {submitting ? 'Submitting...' : 'Submit Task'}
                    </button>
                </div>
            </CardContent>
        </Card>
    );
}