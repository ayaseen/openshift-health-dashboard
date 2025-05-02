import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { getScoreColor, getScoreRating, getHealthAssessment } from '../utils/scoreUtils';

// Custom tooltip formatter for the charts
const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-white p-3 shadow-md rounded-md border border-gray-200">
                <p className="font-medium">{`${payload[0].name}: ${payload[0].value}%`}</p>
                <p className="text-sm text-gray-600">{getScoreRating(payload[0].value)}</p>
            </div>
        );
    }
    return null;
};

const OverviewTab = ({ reportData }) => {
    // Create data for category chart
    const getCategoryData = () => {
        if (!reportData) return [];

        return [
            { name: 'Infrastructure Setup', score: reportData.scoreInfra || 0 },
            { name: 'Policy Governance', score: reportData.scoreGovernance || 0 },
            { name: 'Compliance', score: reportData.scoreCompliance || 0 },
            { name: 'Monitoring', score: reportData.scoreMonitoring || 0 },
            { name: 'Build/Deploy', score: reportData.scoreBuildSecurity || 0 }
        ];
    };

    // Create data for issues chart
    const getIssuesData = () => {
        if (!reportData) return [];

        return [
            {
                name: 'Issues',
                Required: reportData.itemsRequired?.length || 0,
                Recommended: reportData.itemsRecommended?.length || 0,
                Advisory: reportData.itemsAdvisory?.length || 0
            }
        ];
    };

    // Determine if there are any critical issues
    const hasCriticalIssues = (reportData.itemsRequired?.length || 0) > 0;

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-2xl font-bold mb-6">Cluster Health Overview</h2>

            {/* Overall Health Score */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                <div className="bg-white rounded-lg overflow-hidden shadow">
                    <div className="p-6">
                        <h3 className="text-xl font-semibold mb-4">Overall Health Score</h3>
                        <div className="flex items-center justify-center">
                            <div className="w-48 h-48 relative">
                                <ResponsiveContainer width="100%" height="100%">
                                    <PieChart>
                                        <Pie
                                            data={[{ value: reportData.overallScore }, { value: 100 - reportData.overallScore }]}
                                            cx="50%"
                                            cy="50%"
                                            startAngle={90}
                                            endAngle={-270}
                                            innerRadius="60%"
                                            outerRadius="100%"
                                            paddingAngle={0}
                                            dataKey="value"
                                        >
                                            <Cell fill={getScoreColor(reportData.overallScore)} />
                                            <Cell fill="#e5e7eb" />
                                        </Pie>
                                    </PieChart>
                                </ResponsiveContainer>
                                <div className="absolute inset-0 flex items-center justify-center flex-col">
                                    <span className="text-3xl font-bold">{reportData.overallScore}%</span>
                                    <span className="text-sm text-gray-500">{getScoreRating(reportData.overallScore)}</span>
                                </div>
                            </div>
                        </div>
                        <div className="mt-4">
                            <p className="text-gray-700">
                                {getHealthAssessment(reportData.overallScore)}
                            </p>

                            {hasCriticalIssues && (
                                <div className="mt-2 p-2 bg-red-50 border border-red-100 rounded text-sm text-red-700">
                                    <span className="font-medium">Attention needed:</span> {reportData.itemsRequired.length} critical issues require immediate action.
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                <div className="bg-white rounded-lg overflow-hidden shadow md:col-span-2">
                    <div className="p-6">
                        <h3 className="text-xl font-semibold mb-4">Category Scores</h3>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart
                                data={getCategoryData()}
                                layout="vertical"
                                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                            >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis type="number" domain={[0, 100]} />
                                <YAxis type="category" dataKey="name" width={100} />
                                <Tooltip content={<CustomTooltip />} />
                                <Bar dataKey="score" name="Score">
                                    {getCategoryData().map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={getScoreColor(entry.score)} />
                                    ))}
                                </Bar>
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>

            {/* Issues summary */}
            <div className="bg-white rounded-lg overflow-hidden shadow mb-6">
                <div className="p-6">
                    <h3 className="text-xl font-semibold mb-4">Issues Summary</h3>

                    {getIssuesData()[0].Required === 0 &&
                    getIssuesData()[0].Recommended === 0 &&
                    getIssuesData()[0].Advisory === 0 ? (
                        <div className="bg-green-50 p-4 rounded-lg border border-green-100 text-green-800">
                            <p className="font-medium">No issues detected</p>
                            <p className="mt-1 text-sm">Your cluster appears to be following all best practices.</p>
                        </div>
                    ) : (
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart
                                data={getIssuesData()}
                                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                            >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="Required" name="Required Changes" fill="#EF4444" />
                                <Bar dataKey="Recommended" name="Recommended Changes" fill="#F59E0B" />
                                <Bar dataKey="Advisory" name="Advisory Actions" fill="#3B82F6" />
                            </BarChart>
                        </ResponsiveContainer>
                    )}
                </div>
            </div>

            {/* Recent reports info */}
            <div className="bg-blue-50 p-4 rounded-lg border border-blue-100">
                <h3 className="text-lg font-semibold text-blue-800 mb-2">Report Information</h3>
                <p className="text-sm text-blue-700">
                    This report was analyzed on <span className="font-medium">{new Date().toLocaleDateString()}</span>.
                    View the "Remediation Actions" tab for detailed recommendations.
                </p>
            </div>
        </div>
    );
};

export default OverviewTab;