// app/dashboard/src/ExecutiveSummaryDashboard.js
import React from 'react';

const ExecutiveSummaryDashboard = ({ reportData }) => {
    if (!reportData) {
        return null;
    }

    // Helper function to determine color based on score
    const getScoreColor = (score) => {
        if (score >= 90) return '#10B981'; // Green
        if (score >= 70) return '#22C55E'; // Light Green
        if (score >= 50) return '#F59E0B'; // Amber
        return '#EF4444'; // Red
    };

    // Extract categories from report data
    const categories = [
        { name: 'Infrastructure Setup', score: reportData.scoreInfra || 0, description: reportData.infraDescription || '' },
        { name: 'Policy Governance', score: reportData.scoreGovernance || 0, description: reportData.governanceDescription || '' },
        { name: 'Compliance Benchmarking', score: reportData.scoreCompliance || 0, description: reportData.complianceDescription || '' },
        { name: 'Central Monitoring and Logging', score: reportData.scoreMonitoring || 0, description: reportData.monitoringDescription || '' },
        { name: 'Build/Deploy Security', score: reportData.scoreBuildSecurity || 0, description: reportData.buildSecurityDescription || '' }
    ];

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-2xl font-bold mb-6">Executive Summary</h2>

            {/* Overall Health Score */}
            <div className="mb-8">
                <h3 className="text-xl font-semibold mb-2">Overall Health Score</h3>
                <div className="flex items-center">
                    <div className="w-32 h-32 relative">
                        <svg className="w-full h-full">
                            <circle
                                cx="64"
                                cy="64"
                                r="56"
                                fill="none"
                                stroke="#e5e7eb"
                                strokeWidth="10"
                            />
                            <circle
                                cx="64"
                                cy="64"
                                r="56"
                                fill="none"
                                stroke={getScoreColor(reportData.overallScore)}
                                strokeWidth="10"
                                strokeDasharray={352}
                                strokeDashoffset={352 - (352 * reportData.overallScore / 100)}
                                transform="rotate(-90 64 64)"
                            />
                        </svg>
                        <div className="absolute inset-0 flex items-center justify-center">
                            <div className="text-center">
                                <span className="text-3xl font-bold">{reportData.overallScore}%</span>
                                <span className="block text-sm text-gray-500">
                  {reportData.overallScore >= 90 ? 'Excellent' :
                      reportData.overallScore >= 70 ? 'Good' :
                          reportData.overallScore >= 50 ? 'Fair' : 'Poor'}
                </span>
                            </div>
                        </div>
                    </div>
                    <div className="ml-6 flex-1">
                        <p className="text-gray-700">
                            {reportData.overallScore >= 80 ?
                                'Your cluster is well configured and follows best practices.' :
                                reportData.overallScore >= 60 ?
                                    'Your cluster is generally healthy but has some areas needing attention.' :
                                    'Your cluster needs significant improvements to meet best practices.'}
                        </p>
                    </div>
                </div>
            </div>

            {/* Category Scores */}
            <div className="mb-8">
                <h3 className="text-xl font-semibold mb-4">Category Health Assessment</h3>
                <div className="grid grid-cols-1 gap-4">
                    {categories.map((category, index) => (
                        <div key={index} className="bg-gray-50 p-4 rounded-lg">
                            <div className="flex justify-between items-center mb-2">
                                <h4 className="text-lg font-medium">{category.name}</h4>
                                <span
                                    className={`px-3 py-1 rounded-full text-sm font-medium
                    ${category.score >= 90 ? 'bg-green-100 text-green-800' :
                                        category.score >= 70 ? 'bg-green-50 text-green-600' :
                                            category.score >= 50 ? 'bg-yellow-100 text-yellow-800' :
                                                'bg-red-100 text-red-800'}`}
                                >
                  {category.score}%
                </span>
                            </div>
                            <div className="w-full h-4 bg-gray-200 rounded-full overflow-hidden">
                                <div
                                    className="h-full rounded-full"
                                    style={{
                                        width: `${category.score}%`,
                                        backgroundColor: getScoreColor(category.score)
                                    }}
                                ></div>
                            </div>
                            <p className="mt-2 text-gray-600">{category.description}</p>
                        </div>
                    ))}
                </div>
            </div>

            {/* Priority-Based Actions */}
            <div>
                <h3 className="text-xl font-semibold mb-4">Priority-Based Actions</h3>

                {/* Required Changes */}
                <div className="mb-4">
                    <h4 className="text-lg font-medium text-red-600 mb-2">
                        Changes Required ({reportData.itemsRequired?.length || 0})
                    </h4>
                    {reportData.itemsRequired?.length > 0 ? (
                        <ul className="list-disc pl-5 space-y-1">
                            {reportData.itemsRequired.map((item, index) => (
                                <li key={index} className="text-gray-700">{item}</li>
                            ))}
                        </ul>
                    ) : (
                        <p className="text-gray-500">No critical changes required</p>
                    )}
                </div>

                {/* Recommended Changes */}
                <div className="mb-4">
                    <h4 className="text-lg font-medium text-yellow-600 mb-2">
                        Changes Recommended ({reportData.itemsRecommended?.length || 0})
                    </h4>
                    {reportData.itemsRecommended?.length > 0 ? (
                        <ul className="list-disc pl-5 space-y-1">
                            {reportData.itemsRecommended.map((item, index) => (
                                <li key={index} className="text-gray-700">{item}</li>
                            ))}
                        </ul>
                    ) : (
                        <p className="text-gray-500">No recommendations</p>
                    )}
                </div>

                {/* Advisory Actions */}
                <div>
                    <h4 className="text-lg font-medium text-blue-600 mb-2">
                        Advisory Actions ({reportData.itemsAdvisory?.length || 0})
                    </h4>
                    {reportData.itemsAdvisory?.length > 0 ? (
                        <ul className="list-disc pl-5 space-y-1">
                            {reportData.itemsAdvisory.map((item, index) => (
                                <li key={index} className="text-gray-700">{item}</li>
                            ))}
                        </ul>
                    ) : (
                        <p className="text-gray-500">No advisory actions</p>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ExecutiveSummaryDashboard;