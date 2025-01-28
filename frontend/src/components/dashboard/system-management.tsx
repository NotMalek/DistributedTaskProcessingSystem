import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface SystemManagementProps {
    onSystemReset?: () => void;
}

export function SystemManagement({ onSystemReset }: SystemManagementProps) {
    const [error, setError] = useState<string | null>(null);
    const [resetting, setResetting] = useState(false);
    const [showConfirm, setShowConfirm] = useState(false);

    const resetSystem = async () => {
        setResetting(true);
        setError(null);

        try {
            const response = await fetch('http://localhost:8080/api/system/reset', {
                method: 'POST',
            });

            if (!response.ok) {
                throw new Error('Failed to reset system');
            }

            onSystemReset?.();
            setShowConfirm(false);
        } catch (err) {
            setError('Failed to reset system');
        } finally {
            setResetting(false);
        }
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>System Management</CardTitle>
            </CardHeader>
            <CardContent>
                {error && (
                    <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-md flex items-center">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        {error}
                    </div>
                )}

                {!showConfirm ? (
                    <button
                        onClick={() => setShowConfirm(true)}
                        className="w-full py-2 px-4 bg-red-500 text-white rounded-md hover:bg-red-600 flex items-center justify-center"
                    >
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Reset System
                    </button>
                ) : (
                    <div className="space-y-4">
                        <div className="p-4 bg-yellow-50 text-yellow-800 rounded-md">
                            <p className="font-medium">Are you sure?</p>
                            <p className="text-sm mt-1">
                                This will clear all tasks, workers, and system state. This action cannot be undone.
                            </p>
                        </div>
                        <div className="flex space-x-4">
                            <button
                                onClick={() => setShowConfirm(false)}
                                className="flex-1 py-2 px-4 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300"
                                disabled={resetting}
                            >
                                Cancel
                            </button>
                            <button
                                onClick={resetSystem}
                                className="flex-1 py-2 px-4 bg-red-500 text-white rounded-md hover:bg-red-600"
                                disabled={resetting}
                            >
                                {resetting ? 'Resetting...' : 'Confirm Reset'}
                            </button>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}