import React from 'react';

const TabNavigation = ({ activeTab, setActiveTab }) => {
    return (
        <div className="border-b border-gray-200 mb-6">
            <div className="flex -mb-px">
                <button
                    className={`py-4 px-6 font-medium text-sm ${
                        activeTab === 'overview'
                            ? 'border-b-2 border-indigo-500 text-indigo-600'
                            : 'text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                    onClick={() => setActiveTab('overview')}
                >
                    Overview
                </button>
                <button
                    className={`py-4 px-6 font-medium text-sm ${
                        activeTab === 'executive'
                            ? 'border-b-2 border-indigo-500 text-indigo-600'
                            : 'text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                    onClick={() => setActiveTab('executive')}
                >
                    Executive Summary
                </button>
                <button
                    className={`py-4 px-6 font-medium text-sm ${
                        activeTab === 'remediation'
                            ? 'border-b-2 border-indigo-500 text-indigo-600'
                            : 'text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                    onClick={() => setActiveTab('remediation')}
                >
                    Remediation Actions
                </button>
            </div>
        </div>
    );
};

export default TabNavigation;