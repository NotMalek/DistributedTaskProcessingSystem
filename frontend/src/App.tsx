import { Dashboard } from './components/Dashboard/Dashboard';

function App() {
    return (
        <div className="min-h-screen bg-gray-100">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <h1 className="text-3xl font-bold text-gray-900 mb-8">Task Processing System</h1>
                <Dashboard />
            </div>
        </div>
    );
}

export default App;