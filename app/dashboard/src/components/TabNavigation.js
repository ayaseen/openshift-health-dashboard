import React from 'react';

const TabNavigation = ({ activeTab, setActiveTab }) => {
    return (
        <div className="bg-white rounded-lg shadow-sm mb-6 p-1">
            <div className="flex">
                <button
                    className={`py-3 px-6 font-medium text-sm rounded-md transition-all duration-300 ease-in-out ${
                        activeTab === 'overview'
                            ? 'bg-indigo-100 text-indigo-700 font-semibold shadow-sm'
                            : 'text-gray-600 hover:bg-gray-50'
                    }`}
                    onClick={() => setActiveTab('overview')}
                >
                    Overview
                </button>
                <button
                    className={`py-3 px-6 font-medium text-sm rounded-md transition-all duration-300 ease-in-out ${
                        activeTab === 'executive'
                            ? 'bg-indigo-100 text-indigo-700 font-semibold shadow-sm'
                            : 'text-gray-600 hover:bg-gray-50'
                    }`}
                    onClick={() => setActiveTab('executive')}
                >
                    Executive Summary
                </button>
                <button
                    className={`py-3 px-6 font-medium text-sm rounded-md transition-all duration-300 ease-in-out ${
                        activeTab === 'remediation'
                            ? 'bg-indigo-100 text-indigo-700 font-semibold shadow-sm'
                            : 'text-gray-600 hover:bg-gray-50'
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